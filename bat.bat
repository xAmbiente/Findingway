

@echo off
set DISCORD_TOKEN=MTQ4NTg2MTUwNDg1MDA3MTU3Mg.Gg1-Nn.jsRFIXCtBLecJ9Jn1Q8IoQx7H4BExifOU69JXw
set DISCORD_CHANNEL_ID=1485865920273580063
set DATA_CENTRE=Light
set DUTY="The Weapon's Refrain (Ultimate)"

:: Log-Datei für Fehlerprotokollierung
set LOGFILE=findingway_log.txt
echo Starting FindingWay Bot... > %LOGFILE%

start "" /min .\findingway.exe >> %LOGFILE% 2>&1
echo Bot started at %TIME% >> %LOGFILE%

:: Optional: Bei einem Fehler das Protokoll anzeigen
if %ERRORLEVEL% NEQ 0 (
    echo Error occurred during execution, check the log for details.
    type %LOGFILE%
).
