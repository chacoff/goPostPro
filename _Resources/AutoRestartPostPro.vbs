set Service = GetObject ("winmgmts:")
set Shell = WScript.CreateObject("WScript.Shell")

sEXEName = "goPostPro.exe"
sLAUNCHName = "LaunchGoPostPro.exe"
sApplicationPath = "D:\goPostPro\release\"


Shell.CurrentDirectory=sApplicationPath

WScript.Sleep(5000)
'Loop until the system is shutdown or user logs out
while true
 bRunning = false

 'Look for our application. Set the flag bRunning = true
 'If we see that it is running

 for each Process in Service.InstancesOf ("Win32_Process")
  if Process.Name = sEXEName then
   bRunning=true
  ' Shell.SendKeys "{ENTER}"
  End If
 next


'Is our app running?

if (not bRunning) then
 WScript.Sleep(1000)
 'No it is not, launch it
 Shell.Run Chr(34) & sApplicationPath & sLAUNCHName & Chr(34)
 
 Shell.SendKeys "{ENTER}"
' Shell.SendKeys "{ENTER}"
end if




'Sleep a while so we do not hog the cpu
WScript.Sleep(3000)

wend