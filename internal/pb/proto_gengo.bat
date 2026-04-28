@echo off
setlocal

REM Generate Go protobuf files in-place for common/request/push.
REM Requires Docker and image rvolosatovs/protoc.

where docker >nul 2>&1
if errorlevel 1 (
  echo [ERROR] docker not found. Please install Docker Desktop first.
  exit /b 1
)

set "PB_DIR=%~dp0"

docker run --rm -v "%PB_DIR%:/defs" rvolosatovs/protoc ^
  --proto_path=/defs ^
  --go_out=/defs ^
  --go_opt=paths=source_relative ^
  /defs/common.proto /defs/request.proto /defs/push.proto

if errorlevel 1 (
  echo [ERROR] protoc generation failed.
  exit /b 1
)

echo [OK] Generated:
echo   common.pb.go
echo   request.pb.go
echo   push.pb.go
endlocal
exit /b 0