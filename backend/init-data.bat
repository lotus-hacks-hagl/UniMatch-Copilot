@echo off
echo Initializing default data...
go run scripts/seed.go
if %errorlevel% neq 0 (
    echo.
    echo Error: Seeding failed. Please check your DATABASE_URL in .env.
    pause
    exit /b %errorlevel%
)
echo.
echo Data initialization complete.
pause
