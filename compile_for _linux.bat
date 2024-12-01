@echo off
REM 设置环境变量以交叉编译 Linux 版本
set GOOS=linux
set GOARCH=amd64

REM Go 源文件路径（可以是当前目录）
set SRC_PATH=main.go

REM 编译输出文件名
set OUTPUT=myapp_linux

REM 开始编译
echo Compiling %SRC_PATH% for Linux...
go build -o %OUTPUT% %SRC_PATH%

REM 编译结果
if %errorlevel% equ 0 (
    echo Compilation successful! Output: %OUTPUT%
) else (
    echo Compilation failed!
)

REM 清除环境变量
set GOOS=
set GOARCH=
