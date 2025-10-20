@echo off
echo Regenerating Swagger API Documentation...
echo.

cd /d %~dp0

echo Running swag init...
swag init --dir ./cmd/web,./internal/handlers --parseDependency --parseInternal --generalInfo docs.go --output ./cmd/web/docs

if %errorlevel% equ 0 (
    echo.
    echo ✓ Swagger documentation generated successfully!
    echo.
    echo Documentation files created:
    echo   - cmd/web/docs/docs.go
    echo   - cmd/web/docs/swagger.json
    echo   - cmd/web/docs/swagger.yaml
    echo.
    echo Access Swagger UI at: http://localhost:3000/swagger/index.html
    echo.
) else (
    echo.
    echo ✗ Error generating Swagger documentation
    echo.
)

pause

