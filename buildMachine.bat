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
set counter_file=./buildVersion.txt

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

rem generator a build number
if not exist "%counter_file%" (
  echo 1 > "%counter_file%"
)

for /f "usebackq tokens=*" %%a in ("%counter_file%") do set patch_version=%%a
set /a patch_version=%patch_version%+1
echo %patch_version% > "%counter_file%"
setlocal enabledelayedexpansion
set number=%patch_version%
set part1=!number:~0,1!
set part2=!number:~1,1!
set part3=!number:~2!
echo Build Number: %part1%.%part2%.%part3%

rem update XML using xmlstarlet
xmlstarlet ed -L -u /parameters/build/version -v %part1%.%part2%.%part3% ./config/config.xml

rem Build the Go program
go build -o "%target_folder%\%executable_name%"

rem Copy the config file for Build and Debug
copy "config\%config_file%" "%target_folder%\%config_file%"
copy "config\%config_file%" "%config_file%"

echo Build completed.
exit /b 0
endlocal