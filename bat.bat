@echo off
:: Findingway Bot - Windows launcher
:: Set your Discord token here or in your environment
if "%DISCORD_TOKEN%"=="" (
    echo [ERROR] DISCORD_TOKEN is not set. Set it as an environment variable or add it here.
    exit /b 1
)

:: Optional overrides
:: set CONFIG_PATH=config.yaml
:: set DB_PATH=findingway.db

set LOGFILE=findingway_log.txt
echo Starting FindingWay Bot... > %LOGFILE%
echo Started at %DATE% %TIME% >> %LOGFILE%

start "FindingWay" /min .\findingway.exe >> %LOGFILE% 2>&1

if %ERRORLEVEL% NEQ 0 (
    echo Error starting bot - check %LOGFILE% for details.
    type %LOGFILE%
)
