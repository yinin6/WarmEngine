 rmdir /s /q "expTest" 

del /q "main.exe" 

start cmd /k go build -o main.exe main.go

timeout /t 3 /nobreak >nul 

start cmd /k main.exe -n 1 -N 10 -s 0 -S 2

start cmd /k main.exe -n 1 -N 10 -s 1 -S 2

start cmd /k main.exe -n 2 -N 10 -s 0 -S 2

start cmd /k main.exe -n 2 -N 10 -s 1 -S 2

start cmd /k main.exe -n 3 -N 10 -s 0 -S 2

start cmd /k main.exe -n 3 -N 10 -s 1 -S 2

start cmd /k main.exe -n 4 -N 10 -s 0 -S 2

start cmd /k main.exe -n 4 -N 10 -s 1 -S 2

start cmd /k main.exe -n 5 -N 10 -s 0 -S 2

start cmd /k main.exe -n 5 -N 10 -s 1 -S 2

start cmd /k main.exe -n 6 -N 10 -s 0 -S 2

start cmd /k main.exe -n 6 -N 10 -s 1 -S 2

start cmd /k main.exe -n 7 -N 10 -s 0 -S 2

start cmd /k main.exe -n 7 -N 10 -s 1 -S 2

start cmd /k main.exe -n 8 -N 10 -s 0 -S 2

start cmd /k main.exe -n 8 -N 10 -s 1 -S 2

start cmd /k main.exe -n 9 -N 10 -s 0 -S 2

start cmd /k main.exe -n 9 -N 10 -s 1 -S 2

start cmd /k main.exe -n 0 -N 10 -s 0 -S 2

start cmd /k main.exe -n 0 -N 10 -s 1 -S 2

start cmd /k main.exe -c -N 10 -S 2

