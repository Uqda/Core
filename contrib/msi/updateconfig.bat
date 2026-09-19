@echo off
setlocal enabledelayedexpansion
set "DEST=%ALLUSERSPROFILE%\Uqda\uqda.conf"
set "LEGACY=%ALLUSERSPROFILE%\Yggdrasil\yggdrasil.conf"
if exist "%DEST%" (
  uqda.exe -useconffile "%DEST%" -checkconf
  exit /b !errorlevel!
)
if not exist "%ALLUSERSPROFILE%\Uqda" mkdir "%ALLUSERSPROFILE%\Uqda"
if errorlevel 1 exit /b 1
icacls "%ALLUSERSPROFILE%\Uqda" /inheritance:r /grant:r "*S-1-5-18:(OI)(CI)F" "*S-1-5-32-544:(OI)(CI)F" >nul
if errorlevel 1 exit /b 1
set "TEMP_CONFIG=%DEST%.tmp.%RANDOM%"
if exist "%LEGACY%" goto migrate
uqda.exe -genconf > "%TEMP_CONFIG%"
if errorlevel 1 goto failed
goto validate
:migrate
uqda.exe -useconffile "%LEGACY%" -checkconf
if errorlevel 1 goto failed
copy /b "%LEGACY%" "%TEMP_CONFIG%" >nul
if errorlevel 1 goto failed
:validate
uqda.exe -useconffile "%TEMP_CONFIG%" -checkconf
if errorlevel 1 goto failed
if exist "%DEST%" goto failed
move /-Y "%TEMP_CONFIG%" "%DEST%" <nul >nul
if errorlevel 1 goto failed
exit /b 0
:failed
if exist "%TEMP_CONFIG%" del "%TEMP_CONFIG%"
echo Configuration initialization failed; restore or migrate the existing identity manually. 1>&2
exit /b 1
