# Runs jcd_bin, parses and processes its output
# 2020 Jaime de los Hoyos M.

$OldEncoding = [Console]::OutputEncoding
[Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding(1252)

$output = (jcd_bin.exe p $args)

# Results parsing
$output_array = $output.Split("*")
$version =  $output_array[0]
$status = $output_array[1]
$result = $output_array[2]

Write-Host "JVA Directory Navigator $version"

if ($status -ceq "OK") {
   # Success, attempting to switch directory
   Write-Host "Switching current directory to $result ..."
   Set-Location $result
} elseif ($status -ceq "ERROR") {
   # jcd_bin reported an error, print it out
   Write-Host "Error: $result"
} else {
   # Output string does not match expected format
   Write-Host "Error: Output from jcd does not match expected format"
}

[Console]::OutputEncoding = $OldEncoding
