@echo off
setlocal
setlocal enabledelayedexpansion

rem Check if parameters are provided ----------------------------------------
if "%~1" == "" (
    echo Usage: build.bat [executable_name]
    echo Using default: goPostPro.exe
    set executable_name=goPostPro.exe
) else (
    set executable_name=%~1
)

rem Set variables -----------------------------------------------------------
set target_folder=./BuildMachine/release
set config_file=config.xml
set previous_builds_folder=./BuildMachine/previousReleases
set counter_file=./BuildMachine/buildVersion.txt
set icon=./BuildMachine/beam.ico

rem Create target folder if it doesn't exist --------------------------------
if not exist "%previous_builds_folder%" mkdir "%previous_builds_folder%"

rem Generate a build number -------------------------------------------------
if not exist "%counter_file%" (
  echo 1 > "%counter_file%"
)

for /f "usebackq tokens=*" %%a in ("%counter_file%") do set patch_version=%%a

    set previousBuild=%patch_version%

    set previousMajor=!previousBuild:~0,1!
    set previousMinor=!previousBuild:~1,1!
    set previousPatch=!previousBuild:~2!
    echo Previous Build Number: %previousMajor%.%previousMinor%.%previousPatch%
    set previousBuildNumber=%previousMajor%.%previousMinor%.%previousPatch%

    set /a patch_version=%patch_version%+1
    echo %patch_version% > "%counter_file%"
    
    set number=%patch_version%
    set major=!number:~0,1!
    set minor=!number:~1,1!
    set patch=!number:~2!
    echo Build Number: %major%.%minor%.%patch%
    set buildNumber=%major%.%minor%.%patch%


rem Store previous releases -------------------------------------------------
if exist "%target_folder%-%previousBuildNumber%" (
    echo Moving previous release: %previousBuildNumber%
    move "%target_folder%-%previousBuildNumber%" "%previous_builds_folder%"
)

rem update config XML using xmlstarlet --------------------------------------
xmlstarlet ed -L -u /parameters/build/version -v %buildNumber% ./config/config.xml
echo XML update completed with build number: %buildNumber%

rem Build the Go program -----------------------------------------------------
go build -o "%target_folder%-%buildNumber%\%executable_name%"

rem Copy the config file and external libs to complete the release -----------
copy "config\%config_file%" "%target_folder%-%buildNumber%\%config_file%"
copy "_ExternalLibs\TrayRunner\*.*" "%target_folder%-%buildNumber%"
copy "BuildMachine\beam.ico" "%target_folder%-%buildNumber%\beam.ico"

rem Create shortcut -----------------------------------------------------------
call ./_ExternalLibs/ShortcutJS/shortcutJS.bat -linkfile "%target_folder%-%buildNumber%\LaunchgoPostPro.lnk" -target "%~dp0%target_folder%-%buildNumber%\TrayRunner.exe" -linkarguments "goPostPro" -icon "%~dp0%target_folder%-%buildNumber%\beam.ico"
echo Shortcut created successfully.

echo Build completed for release-%buildNumber%.

endlocal
exit /b 0