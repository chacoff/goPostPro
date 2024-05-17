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
set counter_file=buildVersion.txt
set icon=./_Resources/beam.ico

rem Taking the remote buildVersion number -----------------------------------
git fetch

git checkout buildVersion

git pull

    rem Generate a build number ---------------------------------------------
    if not exist "%counter_file%" (
      echo 1 > "%counter_file%"
    )
    
    for /f "usebackq tokens=*" %%a in ("%counter_file%") do set patch_version=%%a
    
        set previousBuild=%patch_version%
    
        set previousMajor=!previousBuild:~0,1!
        set previousMinor=!previousBuild:~1,1!
        set previousPatch=!previousBuild:~2!
        echo Previous Build Number: %previousMajor%.%previousMinor%.%previousPatch%
        
        set /a patch_version=%patch_version%+1
        echo %patch_version% > "%counter_file%"
        
        set number=%patch_version%
        set major=!number:~0,1!
        set minor=!number:~1,1!
        set patch=!number:~2!
        echo Current Build Number: %major%.%minor%.%patch%
        

        rem set build number Variables --------------------------------------
        set previousBuildNumber=%previousMajor%.%previousMinor%.%previousPatch%
        set buildNumber=%major%.%minor%.%patch%

git commit -a -m "updated build version"

git push -u devops buildVersion
if %errorlevel% neq 0 (
    echo Push to devops failed. Attempting to push to origin...
    git push -u origin buildVersion
    if %errorlevel% neq 0 (
        echo Push to 'origin' also failed. Please check your configuration.
        exit /b 1
    ) else (
        echo Push to origin successful!
    )
) else (
    echo Push to devops successful!
)

timeout /t 1

git checkout dev

timeout /t 1

rem Create target folder if it doesn't exist --------------------------------
if not exist "%previous_builds_folder%" mkdir "%previous_builds_folder%"

rem Store previous releases -------------------------------------------------
if exist "%target_folder%-%previousBuildNumber%" (
    echo Moving previous release: %previousBuildNumber%
    move "%target_folder%-%previousBuildNumber%" "%previous_builds_folder%"
)

rem Build the Go program -----------------------------------------------------
go build -o "%target_folder%-%buildNumber%\%executable_name%"

rem Copy the config file and external libs to complete the release -----------
copy "config\%config_file%" "%target_folder%-%buildNumber%\%config_file%"
copy "_ExternalLibs\TrayRunner\*.*" "%target_folder%-%buildNumber%"
copy "_Resources\beam.ico" "%target_folder%-%buildNumber%\beam.ico"

rem update config XML using xmlstarlet --------------------------------------
xmlstarlet ed -L -u /parameters/build/version -v %buildNumber% %target_folder%-%buildNumber%/%config_file%
echo XML update completed with build number: %buildNumber%

rem Create shortcut -----------------------------------------------------------
rem call ./_ExternalLibs/ShortcutJS/shortcutJS.bat -linkfile "%target_folder%-%buildNumber%\LaunchgoPostPro.lnk" -target "%~dp0%target_folder%-%buildNumber%\TrayRunner.exe" -linkarguments "goPostPro" -icon "%~dp0%target_folder%-%buildNumber%\beam.ico"
windres _Resources\icon.rc -O coff -o _Resources\icon.o
gcc _Resources\launchgoPostPro.c _Resources\icon.o -o %target_folder%-%buildNumber%\LaunchGoPostPro.exe
echo Shortcut created successfully.

echo Build completed for release-%buildNumber%.

endlocal
exit /b 0