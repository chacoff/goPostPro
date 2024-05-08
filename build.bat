@echo off
setlocal

rem Check if parameters are provided
if "%~1" == "" (
    echo Usage: build.bat [executable_name]
    echo Using default: goPostPro.exe
    set executable_name=goPostPro.exe
) else (
    set executable_name=%~1
)

rem Set variables
set "datetime=%date:/=-%_%time::=-%"
set target_folder=./Build
set config_file=config.xml
set previous_builds_folder=./Build/previousBuilds

rem Create target folder if it doesn't exist
if not exist "%target_folder%" mkdir "%target_folder%"

rem Move previous builds if they exist
if exist "%target_folder%\%executable_name%" (
    mkdir "%previous_builds_folder%" 2>nul
    for /f "usebackq" %%F in (`dir /b "%target_folder%\%executable_name%"`) do (
        move /y "%target_folder%\%%F" "%previous_builds_folder%\%%~nF-%datetime%%%~xF"
    )
)

rem Move previous config file if it exists
if exist "%target_folder%\%config_file%" (
    mkdir "%previous_builds_folder%" 2>nul
    move /y "%target_folder%\%config_file%" "%previous_builds_folder%\config-%datetime%.xml"
)

rem Build the Go program
go build -o "%target_folder%\%executable_name%"

rem Copy the config file
copy "config\%config_file%" "%target_folder%\%config_file%"

echo Build completed.
exit /b 0
