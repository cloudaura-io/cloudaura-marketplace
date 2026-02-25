$repo = "cloudaura-io/cloudaura-marketplace"
$dest = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:LOCALAPPDATA\conductor-tui" }
$releases = (Invoke-WebRequest -Uri "https://api.github.com/repos/$repo/releases" -UseBasicParsing | ConvertFrom-Json)
$release = $releases | Where-Object { $_.tag_name -like "conductor-tui-v*" } | Select-Object -First 1
if (-not $release) { Write-Error "Could not find a conductor-tui release"; exit 1 }
$tag = $release.tag_name
$url = "https://github.com/$repo/releases/download/$tag/conductor-tui-windows-x64.exe"
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Write-Host "Installing conductor-tui ($tag) to $dest..."
try {
    Invoke-WebRequest -Uri $url -OutFile "$dest\conductor-tui.exe" -UseBasicParsing -ErrorAction Stop
} catch {
    Write-Error "Failed to download ${url}: $_"; exit 1
}
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$dest*") { [Environment]::SetEnvironmentVariable("Path", "$path;$dest", "User") }
Write-Host "Done. Restart terminal and run: conductor-tui"
