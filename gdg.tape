Output website/static/quickstart.gif

Require echo

Set Shell "bash"
Set FontSize 20
Set Width 1600
Set Height 800



Type "gdg tools context list" Sleep 1s  Enter
Sleep 5s
Type "clear" Enter
Type "gdg tools context new demo" Sleep 1s  Enter

# ── phaseAuthType ──
# Auth type: "Basic Authentication" (first option, already selected)
Sleep 2s Enter

# ── phaseBasicCreds (single form: username + password) ──
# Grafana Username
Sleep 2s
Type "admin" Sleep 200ms  Enter

# Grafana Password
Type "admin" Sleep 1000ms  Enter

# ── phaseServerSettings (single form: URL + output path) ──
# Grafana URL
Sleep 2s
Type "http://localhost:3000" Sleep 1s  Enter

# Output Path
Type "test/data" Sleep 1s Enter

# ── phaseFolderScope ──
# Folder scope: "Build a folder allowlist" (second option)
Sleep 2s
Down Enter

# ── phaseFolderMenu — first visit (1 option: Add a folder) ──
Sleep 2s Enter

# ── phaseFolderAdd ──
Sleep 2s
Type "General" Sleep 1s  Enter

# ── phaseFolderMenu — cursor at 0 (Add), 3 options now ──
Sleep 2s Enter

# ── phaseFolderAdd ──
# Folder name: linux/gnu (literal — will be URL-encoded)
Sleep 2s
Type "linux/gnu" Sleep 1s  Enter

# ── phaseFolderMenu — cursor at 0 (Add) ──
Sleep 2s Enter

# ── phaseFolderAdd ──
# Folder name: linux.* (regex metacharacters auto-detected, stored as-is)
Sleep 2s
Type "linux.*" Sleep 1s  Enter

# ── phaseFolderMenu — cursor at 0 (Add). Down to "Test" (position 1) ──
Sleep 2s
Down Sleep 1s Enter

# ── phaseFolderTest ──
Sleep 2s
Type "linux/gnu/foobar" Sleep 1s  Enter

# ── phaseFolderTestResult ──
# "Test another folder name?" default is Yes. Toggle to No.
Sleep 2s
Right Sleep 1s Enter

# ── phaseFolderMenu — cursor at 1 (Test) after returning from test.
# One Down to reach "Done" (position 2). Two Downs would wrap to "Add".
Sleep 2s
Down Enter

# ── phaseConnectionToggle ──
# Configure connection settings? (default No, toggle to Yes)
Sleep 2s
Left Enter

# ── phaseFilterToggle ──
# Configure Connection Filters? (default No, toggle to Yes)
Sleep 2s
Left Sleep 1s Enter

# ── phaseFilterInput (single form: field + regex + inclusive + add more) ──
# Filter Field
Sleep 2s
Type "access" Sleep 1s  Enter

# Filter Regex
Sleep 1s
Type "proxy" Sleep 1s  Enter

# Inclusive filter? (default No, toggle to Yes)
Sleep 1s
Left Sleep 1s Enter

# Add another filter? (default No, skip — submits the form)
Sleep 1s Enter

# ── phaseDefaultCreds (single form: user + password) ──
# Connection Default User
Sleep 2s
Type "ds_admin" Sleep 1s  Enter

# Connection Default Password
Type "ds_pass" Sleep 1s  Enter

# ── phaseCredRuleToggle ──
# Configure Credential Rules? (default No, toggle to Yes)
Sleep 2s
Left Sleep 1s Enter

# ── phaseCredRuleFile ──
# Secure Data File
Sleep 2s
Type "elastic.yaml" Sleep 1s  Enter

# ── phaseCredRuleMatcher (single form: field + regex + add more) ──
# Matching Field
Sleep 2s
Type "type" Sleep 1s  Enter

# Matching Regex
Sleep 1s
Type "elasticsearch" Sleep 1s  Enter

# Add another matcher for this credential file? (default No, skip — submits)
Sleep 1s Enter

# ── phaseCredRuleCreds (single form: user + password + add more) ──
# Credentials for elastic.yaml — User
Sleep 2s
Type "elastic" Sleep 1s  Enter

# Credentials for elastic.yaml — Password
Type "elastic" Sleep 1s  Enter

# Add another credential rule? (default No, skip — submits the form)
Sleep 2s Enter

# ── phaseStorageToggle ──
# Configure cloud storage? (default No, toggle to Yes)
Sleep 2s
Left Enter

# ── phaseStorageProvider ──
# Cloud storage provider: "AWS S3" (second option)
Sleep 2s
Down Sleep 1s Enter

# ── phaseStorageProviderInfo (managed provider docs + OK) ──
Sleep 5s Enter

Sleep 2s
Type "gdg tools context show demo  --all" Sleep 1s  Enter
Sleep 3s
Type "echo encrypted auth using plugin" Sleep 1s Enter
Sleep 3s
Type "cat test/data/secure/auth_demo.yaml" Sleep 1s Enter
Sleep 4s
Type "echo encrypted connection settings using plugin" Sleep 1s Enter
Sleep 3s
Type "cat test/data/secure/default.yaml" Sleep 1s Enter
Sleep 4s
Type "echo encrypted elasticsearch credentials using plugin" Sleep 1s Enter
Sleep 3s
Type "cat test/data/secure/elastic.yaml" Sleep 1s Enter
Sleep 5s
Type "clear" Enter
Type "gdg backup dash list" Sleep 1s  Enter
Sleep 3s
Type "gdg backup dash upload" Sleep 1s  Enter
Sleep 2s
Type "y" Sleep 1s  Enter
Sleep 3s
Type "gdg backup dash list" Sleep 1s  Enter
Sleep 3s
Type "gdg backup dash download" Sleep 1s  Enter
Sleep 3s
Type "gdg backup dash clear" Sleep 1s  Enter
Sleep 2s
Type "y" Sleep 1s  Enter
Sleep 3s
Type "gdg backup dash list" Sleep 1s  Enter
Sleep 3s

Type "gdg tools context delete demo" Sleep 1s  Enter

# Delete credential files? (default No, toggle to Yes)
Sleep 2s
Left Sleep 1s Enter

Sleep 5s
