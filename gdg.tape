Output website/static/quickstart.gif

Require echo

Set Shell "bash"
Set FontSize 20
Set Width 1200
Set Height 600



Type "gdg tools context list" Sleep 500ms  Enter
Sleep 3s
Type "clear" Enter
Type "gdg tools context new demo" Sleep 500ms  Enter

# Auth type: "Basic Authentication" (first option, already selected)
Sleep 500ms Enter

# Grafana Username
Sleep 500ms
Type "admin" Sleep 200ms  Enter

# Grafana Password
Type "admin" Sleep 1000ms  Enter

# Destination Folder
Type "test/data" Enter

# Grafana URL
Type "http://localhost:3000" Sleep 500ms  Enter

# Folder scope: "Build a folder allowlist" (second option)
Sleep 500ms
Down Enter

# Watched Folders: "Add a folder" (only option on first pass)
Sleep 500ms Enter

# Folder name
Type "General" Sleep 500ms  Enter

# Watched Folders: "Add a folder" (first option) to add second folder
Sleep 1s Enter

# Folder name: linux/gnu (literal, no regex metacharacters — will be URL-encoded)
Sleep 500ms
Type "linux/gnu" Sleep 1s  Enter
Sleep 1s

# Watched Folders: "Add a folder" (first option) to add third folder
Sleep 1s Enter

# Folder name: linux.* (contains regex metacharacter *)
Sleep 500ms
Type "linux.*" Sleep 1s  Enter

# "Regex characters detected" confirm: toggle to "Yes, it's a regex"
Sleep 1s
Left Sleep 500ms Enter
Sleep 1s

# Watched Folders: "Test a folder name against current list" (second option)
Sleep 1s
Down Sleep 500ms Enter

# Test folder name: linux/gnu/foobar
Sleep 500ms
Type "linux/gnu/foobar" Sleep 1s  Enter
Sleep 2s

# "Test another folder name?" confirm: No (default)
Sleep 1s Enter

# Watched Folders: select "Done" (third option)
Sleep 1s
Down Down Enter

# Configure connection settings? (default No, toggle to Yes)
Sleep 1s
Left Enter

# Configure Connection Filters? (default No, toggle to Yes)
Sleep 1s
Left Sleep 500ms Enter

# Filter Field
Sleep 500ms
Type "access" Sleep 1s  Enter

# Filter Regex
Sleep 500ms
Type "proxy" Sleep 1s  Enter

# Inclusive filter? (default No, toggle to Yes)
Sleep 1s
Left Sleep 500ms Enter
Sleep 1s

# Test regex? (default No, skip)
Sleep 1s Enter

# Add another filter? (default No, skip)
Sleep 1s Enter

# Connection Default User
Type "ds_admin" Sleep 500ms  Enter

# Connection Default Password
Type "ds_pass" Sleep 500ms  Enter

# Configure Credential Rules? (default No, toggle to Yes)
Sleep 2s
Left Sleep 500ms Enter

# Secure Data File
Sleep 1s
Type "elastic.yaml" Sleep 1s  Enter

# Matching Field
Sleep 1s
Type "type" Sleep 1s  Enter

# Matching Regex
Sleep 1s
Type "elasticsearch" Sleep 1s  Enter

# Test regex? (default No, skip)
Sleep 2s Enter

# Add another matching rule to this credential entry? (default No, skip)
Sleep 2s Enter

# Add another credential rule? (default No, skip)
Sleep 2s Enter

# Configure cloud storage? (default No, toggle to Yes)
Sleep 2s
Left Enter

# Cloud storage provider: "AWS S3" (second option)
Sleep 2s
Down Sleep 500ms Enter
Sleep 3s

Sleep 2s
Type "gdg tools context show demo  --all" Sleep 500ms  Enter
Sleep 3s
Type "echo encrypted auth using plugin" Sleep 500ms Enter
Sleep 3s
Type "cat test/data/secure/auth_demo.yaml" Sleep 500ms Enter
Sleep 4s
Type "echo encrypted connection settings using plugin" Sleep 500ms Enter
Sleep 3s
Type "cat test/data/secure/default.yaml" Sleep 500ms Enter
Sleep 4s
Type "echo encrypted elasticsearch credentials using plugin" Sleep 500ms Enter
Sleep 3s
Type "cat test/data/secure/elastic.yaml" Sleep 500ms Enter
Sleep 5s
Type "clear" Enter
Type "gdg backup dash list" Sleep 500ms  Enter
Sleep 3s
Type "gdg backup dash upload" Sleep 500ms  Enter
Sleep 1s
Type "y" Sleep 500ms  Enter
Sleep 3s
Type "gdg backup dash list" Sleep 500ms  Enter
Sleep 3s
Type "gdg backup dash download" Sleep 500ms  Enter
Sleep 3s
Type "gdg backup dash clear" Sleep 500ms  Enter
Sleep 1s
Type "y" Sleep 500ms  Enter
Sleep 3s
Type "gdg backup dash list" Sleep 500ms  Enter
Sleep 3s

Type "gdg tools context delete demo" Sleep 500ms  Enter

Sleep 5s
