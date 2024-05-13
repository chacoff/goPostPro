@echo off
rem run this script as admin

if not exist goPostPro.exe (
    echo Build the software before installing by running "go build"
    goto :exit
)

sc create goPostProTr2 binpath= "%CD%\goPostPro.exe" start= auto DisplayName= "goPostProTr2"
sc description goPostProTr2 "goPostProTr2"
net start goPostProTr2
sc query goPostProTr2

echo Check example.log

:exit