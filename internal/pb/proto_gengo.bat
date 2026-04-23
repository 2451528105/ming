@echo off
REM 采用docker proto镜像生成go代码
docker run --rm -v %~dp0:/defs rvolosatovs/protoc --proto_path=/defs --go_out=/defs /defs/*.proto

REM 采用docker proto镜像生成cpp代码
REM docker run --rm -v %~dp0:/defs rvolosatovs/protoc --proto_path=/defs --cpp_out=/defs  /defs/*.proto

REM 采用docker proto镜像生成java代码
REM docker run --rm -v %~dp0:/defs rvolosatovs/protoc --proto_path=/defs --java_out=/defs  /defs/*.proto

REM 采用docker proto镜像生成rust代码
REM docker run --rm -v %~dp0:/defs rvolosatovs/protoc --proto_path=/defs --rust_out=/defs  /defs/*.proto

REM 移动代码到当前目录
move /Y %~dp0\pb\*.go %~dp0\
REM 删除临时目录
rmdir /S /Q %~dp0\pb