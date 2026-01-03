$url = "http://bluegreen.localhost"
$i = 0

Write-Host "Starting Zero-Downtime Test on $url..."
Write-Host "Press Ctrl+C to stop."

while ($true) {
    $i++
    try {
        $response = Invoke-WebRequest -Uri $url -Method Head -ErrorAction Stop
        $status = $response.StatusCode
        $color = "Green"
    } catch {
        $status = $_.Exception.Response.StatusCode
        if ($null -eq $status) { $status = "Connection Failed" }
        $color = "Red"
    }

    $timestamp = Get-Date -Format "HH:mm:ss.fff"
    Write-Host "[$timestamp] Request #$i - Status: $status" -ForegroundColor $color
    Start-Sleep -Milliseconds 200
}
