@rem param1 GOPATH
@rem param2 build path
@rem param3 action

@SET GOPATH=%1
@CD %2

@go %3

@rem succeed or failed
@if %errorlevel%==0 (echo %3 success) else (echo %3 failed)