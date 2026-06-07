@echo off
REM Импорт 01_use_case.puml в проект Visual Paradigm (CE 17.x).
REM Перед запуском закройте VP. При ошибке HeadlessException не задавайте HEADLESS=true.

set VP_SCRIPTS=C:\Program Files\Visual Paradigm CE 17.1\scripts
set VPP=C:\Users\super\OneDrive\Документы\VPProjects\SanatoriumUseCase290426.vpp
set PUML=%~dp0vp_import

cd /d "%VP_SCRIPTS%"
call Plugin.bat -project "%VPP%" -pluginid "plugins.plantUML" -pluginargs -action "import" -path "%PUML%"

pause
