.PHONY: build-fe upgrade-fe build-be upgrade-be build upgrade

## fe
build-fe:
	@cd frontend/ && npm run build

upgrade-fe: build-fe
	@sshpass -p 'Hub@aioz1' scp -r ./frontend/dist aioz-ai-hub@10.0.0.154:~/Desktop/dapps/UniMatch-Copilot/frontend/

deploy-frontend:
	@echo "Deploying frontend to Vercel..."
	@cd frontend && npx vercel --prod

## be
build-be:
	@cd backend && make build

upgrade-be: build-be
	@sshpass -p 'Hub@aioz1' scp ./bin/backend-service aioz-ai-hub@10.0.0.154:~/Desktop/dapps/UniMatch-Copilot/bin/

## ai
build-ai:
	@cd ai-service-go && make build

upgrade-ai: build-ai
	@sshpass -p 'Hub@aioz1' scp ./bin/ai-service-go aioz-ai-hub@10.0.0.154:~/Desktop/dapps/UniMatch-Copilot/bin/

## all
build: build-fe build-be build-ai

upgrade: upgrade-fe upgrade-be upgrade-ai