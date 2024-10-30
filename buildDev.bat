@echo off
setlocal
setlocal enabledelayedexpansion

set previousBuildNumber=1.0.059
set number=99994

set major=!number:~0,1!
set minor=!number:~1,1!
set patch=!number:~2!
set patchv=!number:~2,1!
set patch2=!number:~3!

set buildNumber=%major%.%minor%.%patch%
set buildNumber2=%major%,%minor%,%patchv%,%patch2%

set target_folder=./BuildMachine/release
set previous_builds_folder=./BuildMachine/previousReleases
set executable_name=goPostPro.exe
set config_file=config.xml

rem copy version.info.rc --------------------------------
echo Preparing versioninfo and FART
copy "_Resources\versioninfo.rc" "versioninfo.rc"
copy "_Resources\versionlauncher.rc" "versionlauncher.rc"
copy "_Resources\fart.exe" "fart.exe"
copy "_Resources\beam.ico" "beam.ico"
copy "_Resources\golang.ico" "golang.ico"
copy "_Resources\Poppins-SemiBold.ttf" "Poppins-SemiBold.ttf"

rem Update and Generate versioninfo.rc, versioninfo.syso --------------------------------
echo Version: %buildNumber2%
.\fart.exe versioninfo.rc "0,0,0,1" %buildNumber2%
.\fart.exe versionlauncher.rc "0,0,0,1" %buildNumber2%

rem Build versioninfo.syso --------------------------------
windres -i versioninfo.rc -O coff -o versioninfo.syso

rem Create target folder if it doesn't exist --------------------------------
if not exist "%previous_builds_folder%" mkdir "%previous_builds_folder%"

rem Store previous releases -------------------------------------------------
if exist "%target_folder%-%previousBuildNumber%" (
    echo Moving previous release: %previousBuildNumber%
    move "%target_folder%-%previousBuildNumber%" "%previous_builds_folder%"
)

rem Build the Go program -----------------------------------------------------
rem -ldflags "-s -w" removes all debug infos and reduces the binary file size
go build -ldflags "-s -w" -o "%target_folder%-%buildNumber%\%executable_name%"

rem Copy the config file and external libs to complete the release -----------
copy "config\%config_file%" "%target_folder%-%buildNumber%\%config_file%"
copy "_ExternalLibs\TrayRunner\*.*" "%target_folder%-%buildNumber%"
copy "_Resources\beam.ico" "%target_folder%-%buildNumber%\beam.ico"
copy "_Resources\golang.ico" "%target_folder%-%buildNumber%\golang.ico"
copy "_Resources\Poppins-SemiBold.ttf" "%target_folder%-%buildNumber%\Poppins-SemiBold.ttf"

rem update config XML using xmlstarlet --------------------------------------
xmlstarlet ed -L -u /parameters/build/version -v %buildNumber% %target_folder%-%buildNumber%/%config_file%
echo XML update completed with build number: %buildNumber%

rem Create launcher and syso file for launcher ------------------------------
rem windres _Resources\icon.rc -O coff -o _Resources\icon.o
windres -i versionlauncher.rc -O coff -o versionlauncher.syso
gcc _Resources\launchgoPostPro.c versionlauncher.syso -o %target_folder%-%buildNumber%\LaunchGoPostPro.exe
echo Shortcut created successfully.

rem Cleaning -----------------------------------------------------------------
del /F "versioninfo.rc"
del /F "versioninfo.syso"
del /F "versionlauncher.rc"
del /F "versionlauncher.syso"
del /F "fart.exe"
del /F "beam.ico"
del /F "golang.ico"
del /F "Poppins-SemiBold.ttf"

echo Build completed for release-%buildNumber%.

endlocal
exit /b 0