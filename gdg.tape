Output website/static/quickstart.gif

Require echo

Set Shell "bash"
Set FontSize 20
Set Width 1200
Set Height 600



Type "gdg tools context list" Sleep 500ms  Enter
Type "clear" Enter
Type "gdg tools context new demo" Sleep 500ms  Enter
Down@1 Down@1 Down@1 Sleep 500ms Enter
Type "admin" Sleep 200ms  Enter
Type "admin" Sleep 1000ms  Enter
Type "test/data" Enter
Type "http://localhost:3000" Sleep 500ms  Enter
Type "General,linux%2Fgnu/*" Sleep 500ms  Enter
Type "ds_admin" Sleep 500ms  Enter
Type "ds_pass" Sleep 500ms  Enter
Type "gdg tools context show demo" Sleep 500ms  Enter
Type "cat test/data/secure/default.json" Sleep 500ms Enter
Type "echo encrypted auth using plugin" Sleep 500ms Enter
Type "cat test/data/secure/auth_demo.yaml" Sleep 500ms Enter
Type "clear" Sleep 5000ms  Enter
Type "gdg backup dash list" Sleep 500ms  Enter
Type "gdg backup dash upload" Sleep 500ms  Enter
Type "y" Sleep 500ms  Enter
Type "gdg backup dash list" Sleep 500ms  Enter
Type "gdg backup dash download" Sleep 500ms  Enter
Type "gdg backup dash clear" Sleep 500ms  Enter
Type "y" Sleep 500ms  Enter
Type "gdg backup dash list" Sleep 500ms  Enter


Type "gdg tools context delete demo" Sleep 500ms  Enter

Sleep 5s
