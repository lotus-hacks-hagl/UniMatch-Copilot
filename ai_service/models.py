from pydantic import BaseModel, Field
from typing import Optional

class UniversityMetadata(BaseModel):
    name: str = Field(description="Tên trường đại học")
    country: str = Field(description="Quốc gia nơi trường đại học tọa lạc")
    qs_rank: Optional[int] = Field(
        default=None,
        description="Thứ hạng QS của trường, nếu có thể tìm được")
    ielts_min: Optional[float] = Field(
        default=None,
        description="Điểm IELTS tối thiểu, nếu có thể tìm được"
    )
    sat_required: Optional[bool] = Field(
        default=None,
        description="Trường có yêu cầu SAT hay không, nếu có thể tìm được"
    )
    gpa_expectation_normalized: Optional[float] = Field(
        default=None,
        description="Điểm GPA kỳ vọng đã được chuẩn hóa về thang điểm 4.0, nếu có thể tìm được"
    )
    tuition_usd_per_year: Optional[int] = Field(
        default=None,
        description="Học phí mỗi năm tính theo USD, nếu có thể tìm được"
    )
    scholarship_available: Optional[bool] = Field(
        default=None,
        description="Trường có cung cấp học bổng hay không, nếu có thể tìm được"
    )
    scholarship_notes: Optional[str] = Field(
        default=None,
        description="Ghi chú về học bổng nếu có, ví dụ: loại học bổng, giá trị, điều kiện nhận được, nếu có thể tìm được"
    )
    application_deadline: Optional[str] = Field(
        default=None,
        description="Hạn chót nộp hồ sơ (YYYY-MM-DD), nếu có thể tìm được"
    )
    available_majors: Optional[list[str]] = Field(
        default=None,
        description="Danh sách các ngành học có tại trường, nếu có thể tìm được"
    )
    acceptance_rate: Optional[float] = Field(
        default=None,
        description="Tỷ lệ chấp nhận của trường, nếu có thể tìm được"
    )

class CrawlJobRequest(BaseModel):
    job_id: str
    university_id: str                           # BE's UUID
    callback_url: str = ""
    metadata: UniversityMetadata                 # full metadata, nulls = fields to detect


class CrawlFixedFields(BaseModel):
    qs_rank: Optional[int] = Field(
        default=None,
        description="Thứ hạng QS của trường, nếu có thể tìm được"
    )
    ielts_min: Optional[float] = Field(
        default=None,
        description="Điểm IELTS tối thiểu, nếu có thể tìm được"
    )
    sat_required: Optional[bool] = Field(
        default=None,
        description="Trường có yêu cầu SAT hay không, nếu có thể tìm được"
    )
    gpa_expectation_normalized: Optional[float] = Field(
        default=None,
        description="Điểm GPA kỳ vọng đã được chuẩn hóa về thang điểm 4.0, nếu có thể tìm được"
    )
    tuition_usd_per_year: Optional[int] = Field(
        default=None,
        description="Học phí mỗi năm tính theo USD, nếu có thể tìm được"
    )
    scholarship_available: Optional[bool] = Field(
        default=None,
        description="Trường có cung cấp học bổng hay không, nếu có thể tìm được"
    )
    scholarship_notes: Optional[str] = Field(
        default=None,
        description="Ghi chú về học bổng nếu có, ví dụ: loại học bổng, giá trị, điều kiện nhận được, nếu có thể tìm được"
    )
    application_deadline: Optional[str] = Field(
        default=None,
        description="Hạn chót nộp hồ sơ (YYYY-MM-DD), nếu có thể tìm được"
    )
    available_majors: Optional[list[str]] = Field(
        default=None,
        description="Danh sách các ngành học có tại trường, nếu có thể tìm được"
    )
    acceptance_rate: Optional[float] = Field(
        default=None,
        description="Tỷ lệ chấp nhận của trường, nếu có thể tìm được"
    )


class CrawlResult(BaseModel):
    fixed_fields: CrawlFixedFields = Field(
        description="The researched fields found by the agent; leave missing ones as null"
    )
    source_urls: list[str] = Field(
        default_factory=list,
        description="URLs actually used during research"
    )

class AnalyzeInput(BaseModel):
    full_name: str
    gpa_normalized: float
    ielts_overall: Optional[float] = None
    sat_total: Optional[int] = None
    intended_major: str
    budget_usd_per_year: int
    preferred_countries: list[str] = []
    target_intake: str
    scholarship_required: bool = False
    extracurriculars: str = ""
    achievements: str = ""

class AnalyzeJobRequest(BaseModel):
    job_id: str
    case_id: str
    callback_url: str
    input: AnalyzeInput

class ProfileSummary(BaseModel):
    academic_strength: str = Field(
        description="Đánh giá chung về học lực của học viên"
    )
    language_readiness: str = Field(
        description="Đánh giá chung về khả năng ngôn ngữ của học viên"
    )
    budget_band: str = Field(
        description="Đánh giá chung về khả năng tài chính của học viên"
    )
    scholarship_sensitivity: bool = Field(
        description="Nhạy cảm với học bổng"
    )
    strengths: list[str] = Field(
        description="Điểm mạnh của học viên"
    )
    weaknesses: list[str] = Field(
        description="Điểm yếu của học viên"
    )
    risk_tolerance: str = Field(
        description="Mức độ chấp nhận rủi ro của học viên"
    )

class Recommendation(BaseModel):
    university_id: str = Field(
        description="ID của trường"
    )
    university_name: str = Field(
        description="Tên trường"
    )
    tier: str = Field(
        description="safe, match, or reach"
    )
    admission_likelihood_score: int = Field(
        description="Điểm thể hiện khả năng được nhận"
    )
    student_fit_score: int = Field(
        description="Điểm phù hợp của học viên với trường"
    )
    reason: str = Field(
        description="Lý do mà student phù hợp với trường"
    )
    risks: list[str] = Field(
        description="Rủi ro khi apply vào trường"
    )
    improvements: list[str] = Field(
        description="Những điểm cần cải thiện để tăng cơ hội"
    )
    rank_order: int = Field(
        description="Thứ tự ưu tiên của trường"
    )

class AnalyzeResult(BaseModel):
    profile_summary: ProfileSummary = Field(
        description="Tóm tắt profile của học viên"
    )
    recommendations: list[Recommendation] = Field(
        description="Danh sách các trường được đề xuất"
    )
    confidence_score: float = Field(
        description="Điểm thể hiện sự tự tin của AI"
    )
    escalation_reason: Optional[str] = Field(
        description="Lý do cần sự can thiệp của con người. Nếu bạn không thể tìm đủ thông tin để đưa ra quyết định, hãy để lại lý do ở đây"
    )
