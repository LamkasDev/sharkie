@echo off
set "URL=https://www.dropbox.com/scl/fi/cazyyyhu8ds2z161s3a3o/fs.zip?rlkey=yjoq4skh4m7bgm6ku94q4qnst&st=9j0q6th1&dl=1"
set "DIR_NAME=fs"
set "ZIP_NAME=fs.zip"

:: Check if directory exists.
if exist "%DIR_NAME%" (
    echo Directory "%DIR_NAME%" already exists. Skipping.
    pause
    exit /b
)

:: Download the file.
echo Downloading file...
curl -L -o "%ZIP_NAME%" "%URL%"

:: Unzip using PowerShell.
echo Unzipping...
if not exist "%DIR_NAME%" mkdir "%DIR_NAME%"
powershell -Command "Expand-Archive -Force -LiteralPath '%ZIP_NAME%' -DestinationPath '%DIR_NAME%'"

:: Cleanup.
del "%ZIP_NAME%"

echo Done!
pause