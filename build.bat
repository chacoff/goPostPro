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

rem Set colors --------------------------------------------------------------
# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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
        echo -e ${YELLOW}Previous Build Number: %previousMajor%.%previousMinor%.%previousPatch%${NC}
        
        set /a patch_version=%patch_version%+1
        echo %patch_version% > "%counter_file%"
        
        set number=%patch_version%
        set major=!number:~0,1!
        set minor=!number:~1,1!
        set patch=!number:~2!
        echo -e ${YELLOW}Current Build Number: %major%.%minor%.%patch%${NC}
        

        rem set build number Variables --------------------------------------
        set previousBuildNumber=%previousMajor%.%previousMinor%.%previousPatch%
        set buildNumber=%major%.%minor%.%patch%

git commit -a -m "updated build version"

git push -u devops buildVersion
if [ $? -ne 0 ]; then
    echo -e "${RED}Push to 'devops' failed. Attempting to push to 'origin'...${NC}"
    git push -u origin buildVersion
    if [ $? -ne 0 ]; then
        echo -e "${RED}Push to 'origin' also failed. Please check your configuration${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Push to 'origin' successful!${NC}"

timeout /t 1

git checkout dev

timeout /t 1

rem Create target folder if it doesn't exist --------------------------------
if not exist "%previous_builds_folder%" mkdir "%previous_builds_folder%"

rem Store previous releases -------------------------------------------------
if exist "%target_folder%-%previousBuildNumber%" (
    echo -e ${YELLOW}Moving previous release: %previousBuildNumber%${NC}
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