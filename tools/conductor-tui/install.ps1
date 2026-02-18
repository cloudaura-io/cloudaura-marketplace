$repo = "cloudaura-io/conductor-claude-code"
$dest = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\conductor-tui" }
$tag = (Invoke-WebRequest -Uri "https://api.github.com/repos/$repo/releases/latest" -UseBasicParsing | ConvertFrom-Json).tag_name
$url = "https://github.com/$repo/releases/download/$tag/conductor-tui-windows-x64.exe"
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Write-Host "Installing conductor-tui ($tag) to $dest..."
Invoke-WebRequest -Uri $url -OutFile "$dest\conductor-tui.exe" -UseBasicParsing
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$dest*") { [Environment]::SetEnvironmentVariable("Path", "$path;$dest", "User") }
Write-Host "Done. Restart terminal and run: conductor-tui"
