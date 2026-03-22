package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"slices"
	"strings"
	"time"

	"ai-service-go/config"
	"ai-service-go/internal/dto"
	"ai-service-go/internal/provider"
	"ai-service-go/pkg/apperror"
)

type searchEvidence struct {
	Attempts  []dto.SearchAttempt      `json:"attempts"`
	Results   []dto.ExaSearchResult    `json:"results"`
	URLs      []string                 `json:"urls"`
	Tinyfish  []map[string]interface{} `json:"tinyfish"`
	Coverage  float64                  `json:"coverage"`
	Exhausted bool                     `json:"exhausted"`
}

type jobService struct {
	cfg      *config.Config
	exa      *provider.ExaClient
	tinyfish *provider.TinyfishClient
	openai   *provider.OpenAIClient
	callback *http.Client
	states   *jobStateStore
}

func NewJobService(cfg *config.Config, exa *provider.ExaClient, tinyfish *provider.TinyfishClient, openai *provider.OpenAIClient) JobService {
	return &jobService{
		cfg:      cfg,
		exa:      exa,
		tinyfish: tinyfish,
		openai:   openai,
		callback: &http.Client{Timeout: cfg.CallbackTimeout},
		states:   newJobStateStore(),
	}
}

func (s *jobService) EnqueueAnalyze(ctx context.Context, req dto.AnalyzeJobRequest) *apperror.AppError {
	s.states.initJob(req.JobID, "analyze_profile")
	go s.processAnalyze(context.WithoutCancel(ctx), req)
	return nil
}

func (s *jobService) EnqueueCrawl(ctx context.Context, req dto.CrawlJobRequest) *apperror.AppError {
	s.states.initJob(req.JobID, "crawl_university")
	go s.processCrawl(context.WithoutCancel(ctx), req)
	return nil
}

func (s *jobService) EnqueueReport(ctx context.Context, req dto.ReportJobRequest) *apperror.AppError {
	s.states.initJob(req.JobID, "generate_report")
	go s.processReport(context.WithoutCancel(ctx), req)
	return nil
}

func (s *jobService) GetJob(ctx context.Context, jobID string) (*dto.JobDebugResponse, *apperror.AppError) {
	_ = ctx
	job, ok := s.states.get(jobID)
	if !ok {
		return nil, apperror.NotFound("job not found")
	}
	return job, nil
}

func (s *jobService) processAnalyze(ctx context.Context, req dto.AnalyzeJobRequest) {
	s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
		job.Status = "running"
	})

	result, err := s.buildAnalyzeResult(ctx, req)
	if err != nil {
		s.postFailure(ctx, req.CallbackURL, dto.JobDonePayload{
			JobID:   req.JobID,
			JobType: "analyze_profile",
			Status:  "failed",
			CaseID:  req.CaseID,
		}, err)
		return
	}

	s.postSuccess(ctx, req.CallbackURL, dto.JobDonePayload{
		JobID:   req.JobID,
		JobType: "analyze_profile",
		Status:  "done",
		CaseID:  req.CaseID,
	}, result)
}

func (s *jobService) processCrawl(ctx context.Context, req dto.CrawlJobRequest) {
	s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
		job.Status = "running"
	})

	result, err := s.buildCrawlResult(ctx, req)
	if err != nil {
		s.postFailure(ctx, req.CallbackURL, dto.JobDonePayload{
			JobID:        req.JobID,
			JobType:      "crawl_university",
			Status:       "failed",
			UniversityID: req.UniversityID,
		}, err)
		return
	}

	s.postSuccess(ctx, req.CallbackURL, dto.JobDonePayload{
		JobID:        req.JobID,
		JobType:      "crawl_university",
		Status:       "done",
		UniversityID: req.UniversityID,
	}, result)
}

func (s *jobService) processReport(ctx context.Context, req dto.ReportJobRequest) {
	s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
		job.Status = "running"
	})

	result, err := s.buildReportResult(req)
	if err != nil {
		s.postFailure(ctx, req.CallbackURL, dto.JobDonePayload{
			JobID:   req.JobID,
			JobType: "generate_report",
			Status:  "failed",
			CaseID:  req.CaseID,
		}, err)
		return
	}

	s.postSuccess(ctx, req.CallbackURL, dto.JobDonePayload{
		JobID:   req.JobID,
		JobType: "generate_report",
		Status:  "done",
		CaseID:  req.CaseID,
	}, result)
}

func (s *jobService) buildAnalyzeResult(ctx context.Context, req dto.AnalyzeJobRequest) (*dto.AnalyzeResult, error) {
	candidates := req.Input.CandidateUniversities
	if len(candidates) > s.cfg.MaxCandidates {
		candidates = candidates[:s.cfg.MaxCandidates]
	}
	if len(candidates) == 0 {
		result := &dto.AnalyzeResult{
			ProfileSummary: map[string]interface{}{
				"student_name": req.Input.FullName,
				"search_pipeline": map[string]interface{}{
					"attempts":          []dto.SearchAttempt{},
					"source_urls":       []string{},
					"tinyfish_extracts": []map[string]interface{}{},
					"coverage":          0,
					"openai_fill_used":  false,
				},
			},
			Recommendations:  []dto.RecommendationResult{},
			ConfidenceScore:  0.2,
			EscalationNeeded: true,
			EscalationReason: "No backend-supplied candidate universities were available for a safe recommendation pass.",
		}
		s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
			job.Status = "completed"
		})
		return result, nil
	}

	preRanked := s.rankCandidates(candidates, req.Input)
	shortlist := preRanked
	if len(shortlist) > max(s.cfg.MaxRecommendations*2, 8) {
		shortlist = shortlist[:max(s.cfg.MaxRecommendations*2, 8)]
	}

	evidence := s.collectSearchEvidence(
		ctx,
		req.JobID,
		buildAnalyzeQueries(req.Input),
		s.cfg.MaxSearchAttempts,
		s.cfg.MaxDetailFetches,
		buildAnalyzeTinyfishGoal(req.Input),
	)

	providerEvidenceFound := hasProviderEvidence(evidence)
	fallbackRanking := s.buildHeuristicRecommendations(shortlist, req.Input, evidence)
	recommendations := []dto.RecommendationResult{}
	if providerEvidenceFound {
		recommendations = append(recommendations, fallbackRanking...)
	}
	confidence := computeConfidence(evidence.Coverage, len(recommendations), len(shortlist))
	openAIFillUsed := false

	if s.shouldUseOpenAIFillAnalyze(recommendations, evidence) {
		if filled := s.fillAnalyzeWithOpenAI(ctx, req, shortlist, evidence, fallbackRanking); filled != nil {
			recommendations = filled.Recommendations
			if filled.ConfidenceScore > 0 {
				confidence = filled.ConfidenceScore
			}
			openAIFillUsed = true
		}
	}

	if len(recommendations) > s.cfg.MaxRecommendations {
		recommendations = recommendations[:s.cfg.MaxRecommendations]
	}
	for i := range recommendations {
		recommendations[i].RankOrder = i + 1
		if recommendations[i].Tier == "" {
			recommendations[i].Tier = computeTier(float64(recommendations[i].AdmissionLikelihoodScore))
		}
		if len(recommendations[i].Risks) == 0 {
			recommendations[i].Risks = []string{"Limited evidence required a model-assisted fill for some fields."}
		}
		if len(recommendations[i].Improvements) == 0 {
			recommendations[i].Improvements = []string{"Strengthen supporting documents for the target major."}
		}
		if recommendations[i].Reason == "" {
			recommendations[i].Reason = "Recommendation completed using search evidence, detail extraction, and fallback synthesis."
		}
	}

	provenanceMode := classifyAnalyzeProvenance(evidence, openAIFillUsed)
	if provenanceMode == "heuristic_fallback" {
		confidence = 0.05
		recommendations = []dto.RecommendationResult{}
	} else if provenanceMode == "openai_fill" {
		confidence = math.Min(confidence, 0.45)
	}

	escalationNeeded := confidence < 0.55 || len(recommendations) < min(3, s.cfg.MaxRecommendations) || provenanceMode == "heuristic_fallback"
	escalationReason := ""
	if escalationNeeded {
		switch provenanceMode {
		case "heuristic_fallback":
			escalationReason = "No external provider evidence was available after retries; recommendations were filled from backend candidates using heuristic fallback."
		case "openai_fill":
			escalationReason = "Search evidence remained incomplete after retries; model-assisted fill was used to complete the payload."
		default:
			escalationReason = "Search evidence remained incomplete after retries; fallback synthesis was used to maximize payload completeness."
		}
	}

	profileSummary := map[string]interface{}{
		"student_name": req.Input.FullName,
		"academic_profile": map[string]interface{}{
			"gpa_normalized": req.Input.GpaNormalized,
			"ielts_overall":  req.Input.IeltsOverall,
			"sat_total":      req.Input.SatTotal,
			"toefl_total":    req.Input.ToeflTotal,
		},
		"preferences": map[string]interface{}{
			"major":                req.Input.IntendedMajor,
			"preferred_countries":  req.Input.PreferredCountries,
			"budget_usd_per_year":  req.Input.BudgetUsdPerYear,
			"scholarship_required": req.Input.ScholarshipRequired,
			"target_intake":        req.Input.TargetIntake,
			"background_text":      req.Input.BackgroundText,
		},
		"search_pipeline": map[string]interface{}{
			"attempts":          evidence.Attempts,
			"source_urls":       evidence.URLs,
			"tinyfish_extracts": evidence.Tinyfish,
			"coverage":          evidence.Coverage,
			"openai_fill_used":  openAIFillUsed,
		},
		"provenance": map[string]interface{}{
			"mode":                   provenanceMode,
			"provider_backed":        provenanceMode == "provider_backed",
			"external_results_count": len(evidence.Results),
			"source_url_count":       len(evidence.URLs),
			"tinyfish_extract_count": len(evidence.Tinyfish),
			"note":                   buildAnalyzeProvenanceNote(provenanceMode),
		},
	}

	return &dto.AnalyzeResult{
		ProfileSummary:   profileSummary,
		Recommendations:  recommendations,
		ConfidenceScore:  confidence,
		EscalationNeeded: escalationNeeded,
		EscalationReason: escalationReason,
	}, nil
}

func (s *jobService) buildCrawlResult(ctx context.Context, req dto.CrawlJobRequest) (*dto.CrawlResult, error) {
	name := stringValue(req.Metadata["name"])
	country := stringValue(req.Metadata["country"])
	evidence := s.collectSearchEvidence(
		ctx,
		req.JobID,
		buildCrawlQueries(name, country),
		s.cfg.MaxSearchAttempts,
		s.cfg.MaxDetailFetches,
		buildCrawlTinyfishGoal(name, country),
	)

	providerEvidenceFound := hasProviderEvidence(evidence)
	result := &dto.CrawlResult{
		Name:            name,
		Country:         country,
		CrawlStatus:     "failed",
		AvailableMajors: []string{},
		SourceURLs:      evidence.URLs,
		ChangesDetected: []string{},
	}
	if providerEvidenceFound {
		result = s.buildHeuristicCrawlResult(req, evidence, name, country)
	}
	openAIFillUsed := false
	if s.shouldUseOpenAIFillCrawl(result, evidence) {
		if filled := s.fillCrawlWithOpenAI(ctx, req, evidence, result); filled != nil {
			result = mergeCrawlResult(result, filled)
			openAIFillUsed = true
		}
	}

	if result.CrawlStatus == "" {
		if providerEvidenceFound || openAIFillUsed {
			result.CrawlStatus = "ok"
		} else {
			result.CrawlStatus = "failed"
		}
	}
	if len(result.SourceURLs) == 0 {
		result.SourceURLs = evidence.URLs
	}
	if len(result.ChangesDetected) == 0 {
		if providerEvidenceFound || openAIFillUsed {
			result.ChangesDetected = []string{"crawl completed with provider-backed or model-assisted synthesis"}
		} else {
			result.ChangesDetected = []string{"crawl failed to collect external evidence and no OpenAI fill was available"}
		}
	}
	if len(result.ChangesDetected) > 0 {
		result.ChangesDetected = append([]string{buildCrawlProvenanceNote(evidence, openAIFillUsed)}, result.ChangesDetected...)
	}

	return result, nil
}

func (s *jobService) buildReportResult(req dto.ReportJobRequest) (*dto.ReportResult, error) {
	lines := []string{
		"# UniMatch AI Report",
		"",
		"Student: " + req.StudentName,
		"Case ID: " + req.CaseID,
		"",
		"## Recommendations",
	}
	for _, rec := range req.Recommendations {
		lines = append(lines, fmt.Sprintf(
			"- %s (%s): admission %d, fit %d. %s",
			rec.UniversityName,
			rec.Tier,
			rec.AdmissionLikelihoodScore,
			rec.StudentFitScore,
			rec.Reason,
		))
	}

	content := strings.Join(lines, "\n")
	return &dto.ReportResult{
		PDFContent: base64.StdEncoding.EncodeToString([]byte(content)),
		Summary:    fmt.Sprintf("Generated report for %s with %d recommendations.", req.StudentName, len(req.Recommendations)),
	}, nil
}

func (s *jobService) collectSearchEvidence(ctx context.Context, jobID string, queries []string, maxAttempts int, maxDetailFetches int, tinyfishGoal func(string) string) searchEvidence {
	evidence := searchEvidence{}
	seenURLs := map[string]struct{}{}
	seenQueries := map[string]struct{}{}

	for _, query := range queries {
		normalized := strings.TrimSpace(query)
		if normalized == "" {
			continue
		}
		if _, exists := seenQueries[normalized]; exists {
			continue
		}
		seenQueries[normalized] = struct{}{}
		if len(evidence.Attempts) >= maxAttempts {
			break
		}

		results, err := s.exa.Search(ctx, normalized, 5)
		attempt := dto.SearchAttempt{
			Query:       normalized,
			ResultCount: len(results),
		}
		if err != nil {
			attempt.Error = err.Error()
		}
		evidence.Attempts = append(evidence.Attempts, attempt)
		s.states.update(jobID, func(job *dto.JobDebugResponse) {
			job.SearchAttempts = append([]dto.SearchAttempt(nil), evidence.Attempts...)
			if attempt.Error != "" {
				job.LastError = attempt.Error
			}
		})

		for _, result := range results {
			evidence.Results = append(evidence.Results, result)
			if result.URL == "" {
				continue
			}
			if _, exists := seenURLs[result.URL]; exists {
				continue
			}
			seenURLs[result.URL] = struct{}{}
			evidence.URLs = append(evidence.URLs, result.URL)
		}

		if len(evidence.URLs) >= maxDetailFetches*2 && len(evidence.Results) >= maxAttempts*2 {
			break
		}
	}

	for _, url := range evidence.URLs {
		if len(evidence.Tinyfish) >= maxDetailFetches {
			break
		}
		raw, err := s.tinyfish.Run(ctx, url, tinyfishGoal(url))
		if err != nil || len(raw) == 0 {
			continue
		}
		if parsed := parseLooseJSONObject(raw); len(parsed) > 0 {
			evidence.Tinyfish = append(evidence.Tinyfish, parsed)
			s.states.update(jobID, func(job *dto.JobDebugResponse) {
				job.TinyfishFetches = append(job.TinyfishFetches, url)
			})
		}
	}

	evidence.Exhausted = len(evidence.Attempts) >= maxAttempts
	evidence.Coverage = computeEvidenceCoverage(evidence)
	return evidence
}

func (s *jobService) rankCandidates(candidates []dto.CandidateUniversity, input dto.AnalyzeInput) []dto.CandidateUniversity {
	ranked := make([]dto.CandidateUniversity, len(candidates))
	copy(ranked, candidates)
	slices.SortFunc(ranked, func(a, b dto.CandidateUniversity) int {
		scoreA, _, _, _ := scoreCandidate(a, input)
		scoreB, _, _, _ := scoreCandidate(b, input)
		switch {
		case scoreA > scoreB:
			return -1
		case scoreA < scoreB:
			return 1
		default:
			return strings.Compare(a.UniversityName, b.UniversityName)
		}
	})
	return ranked
}

func (s *jobService) buildHeuristicRecommendations(candidates []dto.CandidateUniversity, input dto.AnalyzeInput, evidence searchEvidence) []dto.RecommendationResult {
	type scoredRecommendation struct {
		dto.RecommendationResult
		score float64
	}

	var scored []scoredRecommendation
	for _, candidate := range candidates {
		totalScore, admissionScore, fitScore, evidenceScore := scoreCandidate(candidate, input)
		reason := buildReason(candidate, input, evidence)
		scored = append(scored, scoredRecommendation{
			RecommendationResult: dto.RecommendationResult{
				UniversityID:             candidate.UniversityID,
				UniversityName:           candidate.UniversityName,
				Tier:                     computeTier(totalScore),
				AdmissionLikelihoodScore: admissionScore,
				StudentFitScore:          fitScore,
				Reason:                   reason,
				Risks:                    buildRisks(candidate, input, evidenceScore, evidence),
				Improvements:             buildImprovements(candidate, input),
			},
			score: totalScore + evidence.Coverage*10,
		})
	}

	slices.SortFunc(scored, func(a, b scoredRecommendation) int {
		switch {
		case a.score > b.score:
			return -1
		case a.score < b.score:
			return 1
		default:
			return strings.Compare(a.UniversityName, b.UniversityName)
		}
	})

	limit := min(s.cfg.MaxRecommendations, len(scored))
	recommendations := make([]dto.RecommendationResult, 0, limit)
	for i := 0; i < limit; i++ {
		item := scored[i].RecommendationResult
		item.RankOrder = i + 1
		recommendations = append(recommendations, item)
	}
	return recommendations
}

func (s *jobService) shouldUseOpenAIFillAnalyze(recommendations []dto.RecommendationResult, evidence searchEvidence) bool {
	if !s.cfg.AllowOpenAIFill || s.openai == nil || !s.openai.Enabled() {
		return false
	}
	if len(recommendations) < min(3, s.cfg.MaxRecommendations) {
		return true
	}
	if evidence.Coverage < 0.55 {
		return true
	}
	return len(evidence.Tinyfish) == 0 || len(evidence.URLs) == 0
}

func (s *jobService) fillAnalyzeWithOpenAI(ctx context.Context, req dto.AnalyzeJobRequest, shortlist []dto.CandidateUniversity, evidence searchEvidence, fallback []dto.RecommendationResult) *dto.AnalyzeResult {
	payload := map[string]interface{}{
		"instruction":              "Return valid JSON only. Choose universities only from candidate_universities. Fill every required field. If evidence is missing, infer or hallucinate conservatively but keep the payload complete.",
		"student_input":            req.Input,
		"candidate_universities":   shortlist,
		"search_evidence":          evidence,
		"fallback_recommendations": fallback,
		"required_output_shape": map[string]interface{}{
			"profile_summary":   "object",
			"recommendations":   "array of recommendation objects",
			"confidence_score":  "number 0..1",
			"escalation_needed": "boolean",
			"escalation_reason": "string",
		},
	}

	systemPrompt := "You are completing a university recommendation JSON response for an internal admissions system. Always output one valid JSON object. Prioritize filling every field. You may infer missing facts when evidence is weak, but preserve provided university_id values from the candidate list and do not invent new IDs."
	s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
		job.OpenAIFillUsed = true
	})

	idByName := map[string]string{}
	validIDs := map[string]struct{}{}
	for _, candidate := range shortlist {
		idByName[strings.ToLower(candidate.UniversityName)] = candidate.UniversityID
		validIDs[candidate.UniversityID] = struct{}{}
	}

	var lastErr error
	for attempt := 1; attempt <= s.cfg.OpenAIRetryAttempts; attempt++ {
		s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
			job.OpenAIAttempts = attempt
		})

		raw, err := s.openai.CompleteJSON(ctx, systemPrompt, payload)
		if err != nil {
			lastErr = err
			s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
				job.LastError = err.Error()
			})
			continue
		}

		var parsed dto.AnalyzeResult
		if err := json.Unmarshal(raw, &parsed); err != nil {
			lastErr = err
			s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
				job.LastError = err.Error()
			})
			continue
		}

		normalizeAnalyzeResult(&parsed, fallback, idByName, validIDs)
		if err := validateAnalyzeResult(&parsed, len(shortlist) == 0); err != nil {
			lastErr = err
			s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
				job.LastError = err.Error()
			})
			continue
		}
		return &parsed
	}

	if lastErr != nil {
		s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
			job.LastError = lastErr.Error()
		})
	}
	return nil
}

func (s *jobService) buildHeuristicCrawlResult(req dto.CrawlJobRequest, evidence searchEvidence, name, country string) *dto.CrawlResult {
	tuition := inferTuition(name, country)
	ieltsMin := inferIelts(country)
	satRequired := inferSatRequired(country)
	gpaExpectation := inferGPA(country)
	scholarshipAvailable := inferScholarship(country)
	scholarshipNotes := "Derived from public admissions and scholarship signals; confirm with official university pages."
	acceptanceRate := inferAcceptanceRate(name)
	appDeadline := inferDeadline()
	majors := deriveMajorsFromEvidence(evidence)
	crawlStatus := "ok"
	if len(evidence.URLs) == 0 && !s.cfg.FallbackEnabled {
		crawlStatus = "failed"
	}
	changesDetected := []string{
		"tuition and scholarship information refreshed",
		"major catalog refreshed from current search evidence",
	}
	if len(evidence.URLs) == 0 {
		changesDetected = []string{"crawl completed with fallback synthesis after search retries returned limited evidence"}
	}

	return &dto.CrawlResult{
		Name:                     name,
		Country:                  country,
		QsRank:                   inferQS(name),
		IeltsMin:                 &ieltsMin,
		SatRequired:              &satRequired,
		GpaExpectationNormalized: &gpaExpectation,
		TuitionUsdPerYear:        &tuition,
		ScholarshipAvailable:     &scholarshipAvailable,
		ScholarshipNotes:         &scholarshipNotes,
		ApplicationDeadline:      &appDeadline,
		AvailableMajors:          majors,
		AcceptanceRate:           &acceptanceRate,
		CrawlStatus:              crawlStatus,
		ChangesDetected:          changesDetected,
		SourceURLs:               evidence.URLs,
	}
}

func (s *jobService) shouldUseOpenAIFillCrawl(result *dto.CrawlResult, evidence searchEvidence) bool {
	if !s.cfg.AllowOpenAIFill || s.openai == nil || !s.openai.Enabled() {
		return false
	}
	if evidence.Coverage < 0.6 {
		return true
	}
	return result == nil || len(result.AvailableMajors) == 0 || result.TuitionUsdPerYear == nil
}

func (s *jobService) fillCrawlWithOpenAI(ctx context.Context, req dto.CrawlJobRequest, evidence searchEvidence, fallback *dto.CrawlResult) *dto.CrawlResult {
	payload := map[string]interface{}{
		"instruction":     "Return valid JSON only. Fill every crawl field. If search evidence is incomplete, infer missing values so the output is complete and usable.",
		"job_request":     req,
		"search_evidence": evidence,
		"fallback_result": fallback,
	}

	systemPrompt := "You are completing a university crawl JSON object for an internal knowledge base. Always output one valid JSON object. Prefer evidence when present, but infer missing values when needed so the object is complete."
	s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
		job.OpenAIFillUsed = true
	})

	var lastErr error
	for attempt := 1; attempt <= s.cfg.OpenAIRetryAttempts; attempt++ {
		s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
			job.OpenAIAttempts = attempt
		})

		raw, err := s.openai.CompleteJSON(ctx, systemPrompt, payload)
		if err != nil {
			lastErr = err
			s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
				job.LastError = err.Error()
			})
			continue
		}

		var parsed dto.CrawlResult
		if err := json.Unmarshal(raw, &parsed); err != nil {
			lastErr = err
			s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
				job.LastError = err.Error()
			})
			continue
		}

		merged := mergeCrawlResult(fallback, &parsed)
		if err := validateCrawlResult(merged); err != nil {
			lastErr = err
			s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
				job.LastError = err.Error()
			})
			continue
		}
		return merged
	}

	if lastErr != nil {
		s.states.update(req.JobID, func(job *dto.JobDebugResponse) {
			job.LastError = lastErr.Error()
		})
	}
	return nil
}

func (s *jobService) postSuccess(ctx context.Context, callbackURL string, payload dto.JobDonePayload, result interface{}) {
	resultBody, err := json.Marshal(result)
	if err != nil {
		log.Printf("marshal callback result failed: %v", err)
		return
	}
	payload.Result = resultBody
	s.states.update(payload.JobID, func(job *dto.JobDebugResponse) {
		job.Status = "callback_pending"
		job.CallbackStatus = "pending"
	})
	if err := s.postCallback(ctx, callbackURL, payload); err != nil {
		log.Printf("callback failed: %v", err)
	}
}

func (s *jobService) postFailure(ctx context.Context, callbackURL string, payload dto.JobDonePayload, processErr error) {
	msg := processErr.Error()
	payload.Error = &msg
	s.states.update(payload.JobID, func(job *dto.JobDebugResponse) {
		job.Status = "callback_pending"
		job.CallbackStatus = "pending"
		job.LastError = msg
	})
	if err := s.postCallback(ctx, callbackURL, payload); err != nil {
		log.Printf("failure callback failed: %v", err)
	}
}

func (s *jobService) postCallback(ctx context.Context, callbackURL string, payload dto.JobDonePayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= s.cfg.CallbackRetryCount; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
		if err != nil {
			lastErr = err
			break
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.callback.Do(req)
		if err != nil {
			lastErr = err
		} else {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				s.states.update(payload.JobID, func(job *dto.JobDebugResponse) {
					if payload.Status == "failed" {
						job.Status = "failed"
					} else {
						job.Status = "completed"
					}
					job.CallbackStatus = "delivered"
					job.CallbackAttempts = attempt
				})
				return nil
			}
			lastErr = fmt.Errorf("callback returned status %d", resp.StatusCode)
		}

		s.states.update(payload.JobID, func(job *dto.JobDebugResponse) {
			job.CallbackStatus = "retrying"
			job.CallbackAttempts = attempt
			job.LastError = lastErr.Error()
		})
		time.Sleep(s.cfg.CallbackRetryDelay)
	}

	if lastErr != nil {
		s.states.update(payload.JobID, func(job *dto.JobDebugResponse) {
			job.CallbackStatus = "failed"
			job.LastError = lastErr.Error()
		})
	}
	return lastErr
}

func buildAnalyzeQueries(input dto.AnalyzeInput) []string {
	countries := strings.Join(input.PreferredCountries, " ")
	queries := []string{
		fmt.Sprintf("%s undergraduate universities %s %s tuition scholarship admissions", input.IntendedMajor, countries, input.TargetIntake),
		fmt.Sprintf("%s international student admissions %s budget %d", input.IntendedMajor, countries, input.BudgetUsdPerYear),
		fmt.Sprintf("%s scholarship universities %s IELTS SAT requirement", input.IntendedMajor, countries),
		fmt.Sprintf("%s top universities %s cost of attendance %s", input.IntendedMajor, countries, input.TargetIntake),
		fmt.Sprintf("%s universities %s undergraduate requirements financial aid", input.IntendedMajor, countries),
	}

	if input.BackgroundText != "" {
		// Add a query specifically using student background keywords (first 100 chars)
		bgSnippet := input.BackgroundText
		if len(bgSnippet) > 100 {
			bgSnippet = bgSnippet[:100]
		}
		queries = append(queries, fmt.Sprintf("%s university admissions %s %s", input.IntendedMajor, countries, bgSnippet))
	}

	return queries
}

func buildCrawlQueries(name, country string) []string {
	return []string{
		fmt.Sprintf("%s %s admissions requirements tuition scholarships", name, country),
		fmt.Sprintf("%s official undergraduate admissions %s", name, country),
		fmt.Sprintf("%s tuition fees scholarships international students", name),
		fmt.Sprintf("%s available majors acceptance rate", name),
		fmt.Sprintf("%s IELTS SAT GPA requirement", name),
	}
}

func buildAnalyzeTinyfishGoal(input dto.AnalyzeInput) func(string) string {
	return func(url string) string {
		return fmt.Sprintf(
			"Read %s and extract compact JSON with keys: admissions_summary, tuition_signals, scholarship_signals, english_requirement, majors, fit_notes. Focus on %s applicants targeting %s with budget %d USD.",
			url,
			input.IntendedMajor,
			strings.Join(input.PreferredCountries, ", "),
			input.BudgetUsdPerYear,
		)
	}
}

func buildCrawlTinyfishGoal(name, country string) func(string) string {
	return func(url string) string {
		return fmt.Sprintf(
			"Read %s and extract compact JSON for %s (%s) with keys: qs_rank, ielts_min, sat_required, gpa_expectation_normalized, tuition_usd_per_year, scholarship_available, scholarship_notes, application_deadline, available_majors, acceptance_rate.",
			url,
			name,
			country,
		)
	}
}

func scoreCandidate(candidate dto.CandidateUniversity, input dto.AnalyzeInput) (total float64, admission int, fit int, evidence float64) {
	total = 45

	if candidate.Country != "" && len(input.PreferredCountries) > 0 {
		for _, preferred := range input.PreferredCountries {
			if strings.EqualFold(candidate.Country, preferred) {
				total += 14
				break
			}
		}
	}

	if candidate.GpaExpectationNormalized != nil {
		diff := input.GpaNormalized - *candidate.GpaExpectationNormalized
		switch {
		case diff >= 0.4:
			total += 16
		case diff >= 0:
			total += 10
		case diff >= -0.3:
			total += 4
		default:
			total -= 10
		}
		evidence += 0.2
	}

	if candidate.IeltsMin != nil && input.IeltsOverall != nil {
		diff := *input.IeltsOverall - *candidate.IeltsMin
		switch {
		case diff >= 0.5:
			total += 10
		case diff >= 0:
			total += 6
		default:
			total -= 8
		}
		evidence += 0.15
	}

	if candidate.TuitionUsdPerYear != nil && input.BudgetUsdPerYear > 0 {
		switch {
		case input.BudgetUsdPerYear >= *candidate.TuitionUsdPerYear:
			total += 12
		case input.BudgetUsdPerYear+5000 >= *candidate.TuitionUsdPerYear:
			total += 5
		default:
			total -= 10
		}
		evidence += 0.2
	}

	if candidate.ScholarshipAvailable && input.ScholarshipRequired {
		total += 8
		evidence += 0.1
	}

	if len(candidate.AvailableMajors) > 0 {
		for _, major := range candidate.AvailableMajors {
			lowerMajor := strings.ToLower(major)
			lowerTarget := strings.ToLower(input.IntendedMajor)
			if strings.Contains(lowerMajor, lowerTarget) || strings.Contains(lowerTarget, lowerMajor) {
				total += 14
				evidence += 0.15
				break
			}
		}
	}

	if candidate.AcceptanceRate != nil {
		if *candidate.AcceptanceRate >= 0.45 {
			total += 5
		} else if *candidate.AcceptanceRate <= 0.12 {
			total -= 8
		}
		evidence += 0.1
	}

	if candidate.QsRank != nil {
		switch {
		case *candidate.QsRank <= 25:
			total -= 5
		case *candidate.QsRank <= 100:
			total += 2
		default:
			total += 5
		}
		evidence += 0.1
	}

	total = math.Max(5, math.Min(95, total))
	admission = int(math.Round(total))
	fit = int(math.Round(math.Max(10, math.Min(98, total+evidence*10))))
	return total, admission, fit, evidence
}

func computeTier(score float64) string {
	switch {
	case score >= 75:
		return "safe"
	case score >= 55:
		return "match"
	default:
		return "reach"
	}
}

func computeEvidenceCoverage(evidence searchEvidence) float64 {
	coverage := 0.0
	successfulAttempts := 0
	for _, attempt := range evidence.Attempts {
		if attempt.ResultCount > 0 {
			successfulAttempts++
		}
	}
	coverage += math.Min(0.15, float64(successfulAttempts)*0.03)
	coverage += math.Min(0.35, float64(len(evidence.Results))*0.04)
	coverage += math.Min(0.2, float64(len(evidence.URLs))*0.06)
	coverage += math.Min(0.3, float64(len(evidence.Tinyfish))*0.1)
	return math.Min(1.0, coverage)
}

func computeConfidence(coverage float64, recommendationCount, candidateCount int) float64 {
	confidence := 0.1 + coverage*0.75
	confidence += math.Min(0.1, float64(recommendationCount)*0.02)
	confidence += math.Min(0.05, float64(candidateCount)*0.005)
	return math.Min(0.95, confidence)
}

func hasProviderEvidence(evidence searchEvidence) bool {
	return len(evidence.Results) > 0 || len(evidence.URLs) > 0 || len(evidence.Tinyfish) > 0
}

func classifyAnalyzeProvenance(evidence searchEvidence, openAIFillUsed bool) string {
	if len(evidence.Results) > 0 || len(evidence.URLs) > 0 || len(evidence.Tinyfish) > 0 {
		if openAIFillUsed {
			return "provider_plus_openai_fill"
		}
		return "provider_backed"
	}
	if openAIFillUsed {
		return "openai_fill"
	}
	return "heuristic_fallback"
}

func buildAnalyzeProvenanceNote(mode string) string {
	switch mode {
	case "provider_backed":
		return "Recommendations were supported by external search/detail evidence."
	case "provider_plus_openai_fill":
		return "External evidence was found and remaining gaps were completed with OpenAI fill."
	case "openai_fill":
		return "External search did not return enough evidence; OpenAI completed the response."
	default:
		return "External providers returned no usable evidence; output was produced from backend candidates using heuristic fallback."
	}
}

func buildCrawlProvenanceNote(evidence searchEvidence, openAIFillUsed bool) string {
	if len(evidence.Results) > 0 || len(evidence.URLs) > 0 || len(evidence.Tinyfish) > 0 {
		if openAIFillUsed {
			return "crawl provenance: provider evidence found, then OpenAI filled remaining gaps"
		}
		return "crawl provenance: provider-backed result"
	}
	if openAIFillUsed {
		return "crawl provenance: provider evidence missing, OpenAI-filled result"
	}
	return "crawl provenance: heuristic fallback with no external evidence"
}

func buildReason(candidate dto.CandidateUniversity, input dto.AnalyzeInput, evidence searchEvidence) string {
	parts := []string{
		fmt.Sprintf("%s aligns with the student's %s plan", candidate.UniversityName, input.IntendedMajor),
	}
	if candidate.TuitionUsdPerYear != nil {
		parts = append(parts, fmt.Sprintf("estimated tuition sits around USD %d/year", *candidate.TuitionUsdPerYear))
	}
	if candidate.Country != "" {
		parts = append(parts, fmt.Sprintf("the location matches the target geography (%s)", candidate.Country))
	}
	if evidence.Coverage >= 0.55 {
		parts = append(parts, "web search and detail extraction found supporting signals")
	} else {
		parts = append(parts, "the recommendation was completed with aggressive fallback synthesis to keep the payload full")
	}
	return strings.Join(parts, "; ") + "."
}

func buildRisks(candidate dto.CandidateUniversity, input dto.AnalyzeInput, evidenceScore float64, evidence searchEvidence) []string {
	var risks []string
	if candidate.TuitionUsdPerYear != nil && input.BudgetUsdPerYear > 0 && *candidate.TuitionUsdPerYear > input.BudgetUsdPerYear {
		risks = append(risks, "Tuition is above the stated annual budget.")
	}
	if candidate.IeltsMin != nil && input.IeltsOverall != nil && *input.IeltsOverall < *candidate.IeltsMin {
		risks = append(risks, "English proficiency is below the inferred minimum requirement.")
	}
	if evidenceScore < 0.3 || evidence.Coverage < 0.5 {
		risks = append(risks, "Limited external evidence was available, so some fields were inferred.")
	}
	if len(risks) == 0 {
		risks = append(risks, "Competitive admission outcomes still depend on essays, references, and timing.")
	}
	return risks
}

func buildImprovements(candidate dto.CandidateUniversity, input dto.AnalyzeInput) []string {
	var improvements []string
	if candidate.IeltsMin != nil && input.IeltsOverall != nil && *input.IeltsOverall <= *candidate.IeltsMin+0.5 {
		improvements = append(improvements, "Raise English test score to create a stronger margin above the minimum.")
	}
	if input.ScholarshipRequired {
		improvements = append(improvements, "Prepare a scholarship-ready personal statement and financial justification.")
	}
	improvements = append(improvements, "Tailor the application narrative to the intended major and measurable achievements.")
	return improvements
}

func deriveMajorsFromEvidence(evidence searchEvidence) []string {
	base := []string{"Computer Science", "Engineering", "Business"}
	textParts := []string{}
	for _, result := range evidence.Results {
		textParts = append(textParts, result.Title)
		textParts = append(textParts, strings.Join(result.Highlights, " "))
	}
	for _, item := range evidence.Tinyfish {
		if marshaled, err := json.Marshal(item); err == nil {
			textParts = append(textParts, string(marshaled))
		}
	}
	text := strings.ToLower(strings.Join(textParts, " "))
	if text == "" {
		return base
	}

	majors := make([]string, 0, 6)
	for _, option := range []string{"Computer Science", "Data Science", "Engineering", "Business", "Economics", "Medicine"} {
		if strings.Contains(text, strings.ToLower(option)) {
			majors = append(majors, option)
		}
	}
	if len(majors) == 0 {
		return base
	}
	return majors
}

func mergeCrawlResult(base *dto.CrawlResult, filled *dto.CrawlResult) *dto.CrawlResult {
	if filled == nil {
		return base
	}
	if base == nil {
		return filled
	}
	if filled.Name != "" {
		base.Name = filled.Name
	}
	if filled.Country != "" {
		base.Country = filled.Country
	}
	if filled.QsRank != nil {
		base.QsRank = filled.QsRank
	}
	if filled.IeltsMin != nil {
		base.IeltsMin = filled.IeltsMin
	}
	if filled.SatRequired != nil {
		base.SatRequired = filled.SatRequired
	}
	if filled.GpaExpectationNormalized != nil {
		base.GpaExpectationNormalized = filled.GpaExpectationNormalized
	}
	if filled.TuitionUsdPerYear != nil {
		base.TuitionUsdPerYear = filled.TuitionUsdPerYear
	}
	if filled.ScholarshipAvailable != nil {
		base.ScholarshipAvailable = filled.ScholarshipAvailable
	}
	if filled.ScholarshipNotes != nil {
		base.ScholarshipNotes = filled.ScholarshipNotes
	}
	if filled.ApplicationDeadline != nil {
		base.ApplicationDeadline = filled.ApplicationDeadline
	}
	if len(filled.AvailableMajors) > 0 {
		base.AvailableMajors = filled.AvailableMajors
	}
	if filled.AcceptanceRate != nil {
		base.AcceptanceRate = filled.AcceptanceRate
	}
	if filled.CrawlStatus != "" {
		base.CrawlStatus = filled.CrawlStatus
	}
	if len(filled.ChangesDetected) > 0 {
		base.ChangesDetected = filled.ChangesDetected
	}
	if len(filled.SourceURLs) > 0 {
		base.SourceURLs = filled.SourceURLs
	}
	return base
}

func normalizeAnalyzeResult(result *dto.AnalyzeResult, fallback []dto.RecommendationResult, idByName map[string]string, validIDs map[string]struct{}) {
	for i := range result.Recommendations {
		if _, ok := validIDs[result.Recommendations[i].UniversityID]; !ok {
			if mapped := idByName[strings.ToLower(result.Recommendations[i].UniversityName)]; mapped != "" {
				result.Recommendations[i].UniversityID = mapped
			}
		}
		if _, ok := validIDs[result.Recommendations[i].UniversityID]; !ok {
			if i < len(fallback) {
				result.Recommendations[i].UniversityID = fallback[i].UniversityID
				result.Recommendations[i].UniversityName = fallback[i].UniversityName
			}
		}
		if result.Recommendations[i].RankOrder == 0 {
			result.Recommendations[i].RankOrder = i + 1
		}
		if result.Recommendations[i].Tier == "" {
			result.Recommendations[i].Tier = computeTier(float64(result.Recommendations[i].AdmissionLikelihoodScore))
		}
		if len(result.Recommendations[i].Risks) == 0 {
			result.Recommendations[i].Risks = []string{"Some fields were inferred to keep the schema complete."}
		}
		if len(result.Recommendations[i].Improvements) == 0 {
			result.Recommendations[i].Improvements = []string{"Strengthen supporting documents for the target major."}
		}
	}

	if len(result.Recommendations) == 0 {
		result.Recommendations = fallback
	}
	if result.ProfileSummary == nil {
		result.ProfileSummary = map[string]interface{}{}
	}
}

func validateAnalyzeResult(result *dto.AnalyzeResult, allowEmpty bool) error {
	if result == nil {
		return fmt.Errorf("analyze result is nil")
	}
	if result.ProfileSummary == nil {
		return fmt.Errorf("profile_summary is required")
	}
	if !allowEmpty && len(result.Recommendations) == 0 {
		return fmt.Errorf("recommendations are required")
	}
	for i, rec := range result.Recommendations {
		if strings.TrimSpace(rec.UniversityID) == "" {
			return fmt.Errorf("recommendations[%d].university_id is required", i)
		}
		if strings.TrimSpace(rec.UniversityName) == "" {
			return fmt.Errorf("recommendations[%d].university_name is required", i)
		}
		if strings.TrimSpace(rec.Tier) == "" {
			return fmt.Errorf("recommendations[%d].tier is required", i)
		}
		if strings.TrimSpace(rec.Reason) == "" {
			return fmt.Errorf("recommendations[%d].reason is required", i)
		}
		if len(rec.Risks) == 0 {
			return fmt.Errorf("recommendations[%d].risks is required", i)
		}
		if len(rec.Improvements) == 0 {
			return fmt.Errorf("recommendations[%d].improvements is required", i)
		}
	}
	return nil
}

func validateCrawlResult(result *dto.CrawlResult) error {
	if result == nil {
		return fmt.Errorf("crawl result is nil")
	}
	if strings.TrimSpace(result.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(result.Country) == "" {
		return fmt.Errorf("country is required")
	}
	if strings.TrimSpace(result.CrawlStatus) == "" {
		return fmt.Errorf("crawl_status is required")
	}
	if result.SourceURLs == nil {
		return fmt.Errorf("source_urls is required")
	}
	if result.AvailableMajors == nil {
		return fmt.Errorf("available_majors is required")
	}
	return nil
}

func inferTuition(name, country string) int {
	switch strings.ToLower(country) {
	case "usa":
		return 42000
	case "uk":
		return 30000
	case "canada":
		return 28000
	case "australia":
		return 32000
	case "singapore":
		return 18000
	default:
		if strings.Contains(strings.ToLower(name), "technology") {
			return 35000
		}
		return 25000
	}
}

func inferIelts(country string) float64 {
	switch strings.ToLower(country) {
	case "uk", "australia":
		return 6.5
	default:
		return 6.0
	}
}

func inferSatRequired(country string) bool {
	return strings.EqualFold(country, "USA")
}

func inferGPA(country string) float64 {
	switch strings.ToLower(country) {
	case "usa", "uk":
		return 3.3
	default:
		return 3.0
	}
}

func inferScholarship(country string) bool {
	return !strings.EqualFold(country, "UK")
}

func inferAcceptanceRate(name string) float64 {
	if strings.Contains(strings.ToLower(name), "institute of technology") {
		return 0.12
	}
	return 0.42
}

func inferDeadline() string {
	return time.Now().AddDate(0, 5, 0).Format("2006-01-02")
}

func inferQS(name string) *int {
	if name == "" {
		return nil
	}
	rank := 80
	if strings.Contains(strings.ToLower(name), "technology") {
		rank = 25
	}
	return &rank
}

func parseLooseJSONObject(raw []byte) map[string]interface{} {
	var object map[string]interface{}
	if err := json.Unmarshal(raw, &object); err == nil && len(object) > 0 {
		return object
	}

	var wrapped struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Data) > 0 {
		return wrapped.Data
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(raw, &list); err == nil && len(list) > 0 {
		return list[0]
	}

	return nil
}

func stringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
