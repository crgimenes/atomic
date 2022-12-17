if (getEnv("LANG") == "") then
    setOutputMode("CP850")
    -- setOutputMode("CP437")
    -- setOutputMode("UTF-8") -- default
end

-- setOutputDelay(1) -- 1ms delay between each character

cls()
write("\027[0;0H") -- move cursor to 0,0H
write("\r\nthis is a test write to client instance\r\n")
write(" ██████╗██████╗  ██████╗    ███████╗████████╗██╗   ██████╗ ██████╗\r\n")
write("██╔════╝██╔══██╗██╔════╝    ██╔════╝╚══██╔══╝██║   ██╔══██╗██╔══██╗\r\n")
write("██║     ██████╔╝██║  ███╗   █████╗     ██║   ██║   ██████╔╝██████╔╝\r\n")
write("██║     ██╔══██╗██║   ██║   ██╔══╝     ██║   ██║   ██╔══██╗██╔══██╗\r\n")
write("╚██████╗██║  ██║╚██████╔╝██╗███████╗   ██║   ██║██╗██████╔╝██║  ██║\r\n")
write(" ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝\r\n")
write("crg@crg.eti.br @crgimenes\r\n")
write("██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██\r\n")
write("\r\n")

writeFromASCII("nonfree/squiddy.ans")
write("\r\n")

Str = ""

write("output test with accented characters: áéíóú äëïöü ãõ ç\r\n")

-- setOutputDelay(0) -- no delay between each character

function ExecTest()
    exec("zsh")
    cls()
    menu()
end

function ExitConnection()
    print("quit user")
    cls()
    write("\r\nbye!\r\n")
    quit()
end

function TestPrint()
    write("Você digitou: ")
    write(Str)
    write("\r\n")
    inlineImagesProtocol("nonfree/crg.png")
    write("\r\n")
end

Echo = false
function ToggleEcho()
    Echo = not Echo
    setEcho(Echo)
end

clockAux = ""
clockInt = 0
function Clock()
    -- setOutputDelay(0)
    write("\027[s") -- save cursor position ANSI.SYS
    write("\027[?25l") -- hide cursor

    write("\0277\27[0;0H") -- save cursor (DEC) and move cursor to 0,0
    write(os.date('%Y-%m-%d %H:%M:%S UTC'))
    -- os.date("%Y-%m-%dT%H:%m:%S.000 %z"
    write("\0278") -- restore cursor (DEC)
    -- rmTrigger("clock")

    write("\027[?25h") -- show cursor
    write("\027[u") -- restore cursor position ANSI.SYS
    -- setOutputDelay(1)
end

function menu()
    write("[1] toggle echo on/off\r\n")
    write("[2] print test string\r\n")
    write("[3] quit\r\n")
    write("[4] zsh\r\n")
    write("[5] Show Term parameters\r\n")
    write("choose an option\r\n")
end

function ShowTerm()
    write("\r\nTERM:")
    local lang = getEnv("LANG")
    write(lang)
    write("\r\nOUTPUT MODE:")
    write(getOutputMode())
    write("\r\n")
    menu()
end

function ShowSquiddy()
    writeFromASCII("nonfree/squiddy.ans")
    write("\r\n")
    menu()
end

menu()
timer("clock", 2100, Clock)

trigger("1", ToggleEcho)
trigger("2", TestPrint)
trigger("3", ExitConnection)
trigger("4", ExecTest)
trigger("5", ShowTerm)
trigger("6", ShowSquiddy)



write("\r\nLANG = ")
local lang = getEnv("LANG")
if (lang == "") then
    write("empty")
else
    write(lang)
end
write("\r\n")

while true do
    write("\nenter a string:")
    Str = getField()
    write("\r\n")
    write("[")
    write(Str)
    write("]")
    write("\r\n")
end
-- quit()
