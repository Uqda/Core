# Signs all *.msi files in the current directory using Authenticode (signtool).
# Expects environment variables:
#   WINDOWS_CODESIGN_PFX_BASE64 - PFX file bytes, base64-encoded
#   WINDOWS_CODESIGN_PASSWORD    - PFX password
#
# If WINDOWS_CODESIGN_PFX_BASE64 is empty, exits 0 with a warning (unsigned MSI).

$ErrorActionPreference = "Stop"

if (-not $env:WINDOWS_CODESIGN_PFX_BASE64) {
    Write-Host "::warning::WINDOWS_CODESIGN_PFX_BASE64 is not set — MSI will not be signed. Add repository secrets to enable Authenticode (see SECURITY.md)."
    exit 0
}
if (-not $env:WINDOWS_CODESIGN_PASSWORD) {
    Write-Error "WINDOWS_CODESIGN_PASSWORD must be set when WINDOWS_CODESIGN_PFX_BASE64 is set."
    exit 1
}

$kitsRoot = Join-Path ${env:ProgramFiles(x86)} "Windows Kits\10\bin"
if (-not (Test-Path $kitsRoot)) {
    Write-Error "Windows SDK (signtool) not found under $kitsRoot"
    exit 1
}

$preferArch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "x64" }
$candidates = Get-ChildItem -Path $kitsRoot -Recurse -Filter "signtool.exe" -ErrorAction SilentlyContinue |
    Where-Object { $_.Directory.Name -eq $preferArch }
$signtool = $candidates | Sort-Object { $_.FullName } -Descending | Select-Object -First 1
if (-not $signtool) {
    $signtool = Get-ChildItem -Path $kitsRoot -Recurse -Filter "signtool.exe" -ErrorAction SilentlyContinue |
        Where-Object { $_.Directory.Name -match "^(x64|arm64)$" } |
        Sort-Object FullName -Descending |
        Select-Object -First 1
}
if (-not $signtool) {
    Write-Error "signtool.exe not found under Windows Kits"
    exit 1
}

Write-Host "Using $($signtool.FullName)"

$tempRoot = if ($env:RUNNER_TEMP) { $env:RUNNER_TEMP } else { $env:TEMP }
$pfxPath = Join-Path $tempRoot "uqda-codesign.pfx"
try {
    try {
        [IO.File]::WriteAllBytes($pfxPath, [Convert]::FromBase64String($env:WINDOWS_CODESIGN_PFX_BASE64))
    } catch {
        Write-Error "Failed to decode WINDOWS_CODESIGN_PFX_BASE64: $_"
        exit 1
    }

    $msis = Get-ChildItem -Path "." -Filter "*.msi" -File
    if ($msis.Count -eq 0) {
        Write-Error "No .msi files in current directory."
        exit 1
    }

    $tsa = "http://timestamp.digicert.com"
    foreach ($msi in $msis) {
        $signArgs = @(
            "sign",
            "/f", $pfxPath,
            "/p", $env:WINDOWS_CODESIGN_PASSWORD,
            "/fd", "SHA256",
            "/td", "SHA256",
            "/tr", $tsa,
            $msi.FullName
        )
        & $signtool.FullName @signArgs
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

        & $signtool.FullName @("verify", "/pa", "/v", $msi.FullName)
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

        Write-Host "Signed and verified: $($msi.Name)"
    }
    Write-Host "Authenticode signing completed."
} finally {
    Remove-Item -LiteralPath $pfxPath -Force -ErrorAction SilentlyContinue
}
