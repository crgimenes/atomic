Term = require("term")
require "main_menu"

if (getEnv("LANG") == "") then
    Term.setOutputMode("CP850")
    -- Term.setOutputMode("CP437")
    -- Term.setOutputMode("UTF-8") -- default
end

Printf = function(s, ...)
    return io.write(s:format(...))
end

function Menu()
    clearTriggers()
    timer("clock", 1100, Clock)
    trigger("1", Main_menu)
    trigger("2", TestPrint)
    trigger("3", ExitConnection)
    trigger("4", ExecTest)
    trigger("5", ShowTerm)
    trigger("6", ShowSquiddy)
    trigger("7", ShowUsers)
    Term.cls()

    Term.print(5, 8, "1 Main menu")
    Term.print(6, 8, "2 print test string")
    Term.print(7, 8, "3 quit")
    Term.print(8, 8, "4 zsh")
    Term.print(9, 8, "5 Show Term parameters")
    Term.print(10, 8, "6 Show Squiddy")
    Term.print(11, 8, "7 Show Users")
    Term.drawBox(4, 7, 25, 9)

    -- Term.write("choose an option\r\n")
    Term.print(8, 13, "option: ")


    local w, h = Term.getSize()
    Term.drawBox(3, 3, w - 3, h - 3)


end

-- Term.setOutputDelay(1) -- 1ms delay between each character

Term.write("\r\n")
Term.write("\r\nthis is a test write to client instance\r\n")
Term.write(" ██████╗██████╗  ██████╗    ███████╗████████╗██╗   ██████╗ ██████╗\r\n")
Term.write("██╔════╝██╔══██╗██╔════╝    ██╔════╝╚══██╔══╝██║   ██╔══██╗██╔══██╗\r\n")
Term.write("██║     ██████╔╝██║  ███╗   █████╗     ██║   ██║   ██████╔╝██████╔╝\r\n")
Term.write("██║     ██╔══██╗██║   ██║   ██╔══╝     ██║   ██║   ██╔══██╗██╔══██╗\r\n")
Term.write("╚██████╗██║  ██║╚██████╔╝██╗███████╗   ██║   ██║██╗██████╔╝██║  ██║\r\n")
Term.write(" ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝\r\n")
Term.write("crg@crg.eti.br @crgimenes\r\n")
Term.write("██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██\r\n")
Term.write("\r\n")


if (fileExists("nonfree/squiddy.ans")) then
    Term.writeFromASCII("nonfree/squiddy.ans")
    Term.write("\r\n")
end

Str = ""

Term.write("output test with accented characters: áéíóú äëïöü ãõ ç\r\n")

-- Term.setOutputDelay(0) -- no delay between each character

function ExecTest()
    exec("zsh")
    Term.cls()
    Menu()
end

function ExitConnection()
    local u = getUser()
    logf("quit user %s", u.nickname)
    Term.cls()
    Term.write("\r\nbye!\r\n")
    quit()
end

function TestPrint()
    Term.write("\n\r\n\rtest: ")
    Term.write(Str)
    -- write("\r\n")
    -- Term.inlineImagesProtocol("nonfree/crg.png")
    Term.write("\r\n\r\n")
end

Echo = false
function ToggleEcho()
    Echo = not Echo
    Term.setEcho(Echo)
end

clockAux = ""
clockInt = 0
function Clock()
    -- Term.setOutputDelay(0)
    Term.write("\027[s") -- save cursor position ANSI.SYS
    Term.write("\027[?25l") -- hide cursor

    Term.write("\0277\27[0;0H") -- save cursor (DEC) and move cursor to 0,0
    Term.write(os.date('%Y-%m-%d %H:%M:%S'))
    -- os.date("%Y-%m-%dT%H:%m:%S.000 %z"
    Term.write("\0278") -- restore cursor (DEC)
    -- rmTrigger("clock")

    Term.write("\027[?25h") -- show cursor
    Term.write("\027[u") -- restore cursor position ANSI.SYS
    -- Term.setOutputDelay(1)
end

function ShowTerm()
    Term.write("\r\n\r\nTERM:")
    local lang = getEnv("LANG")
    Term.write(lang)
    Term.write("\r\nOUTPUT MODE:")
    Term.write(Term.getOutputMode())
    Term.write("\r\n\r\n")
    Menu()
end

function ShowSquiddy()
    Term.writeFromASCII("nonfree/squiddy.ans")
    Term.write("\r\n")
end

function ShowUsers()

    Term.write("\r\n")
    Term.write("user nick: ")
    local u = getUser()
    Term.write(u.nickname)
    Term.write("\r\n")

    Term.write("user:")
    Term.write("\r\n")

    for k, v in pairs(u) do
        Term.write("\t")
        Term.write(k)
        Term.write(":")
        Term.write(v)
        Term.write("\r\n")
    end
    Term.write("\r\n")

    Menu()
end

Term.write("\r\nLANG = ")
local lang = getEnv("LANG")
if (lang == "") then
    Term.write("empty")
else
    Term.write(lang)
end
Term.write("\r\n")


ShowUsers()

while true do
    -- Term.write("\nenter a string:")
    Term.print(25, 25, "enter a string: ")
    Str = Term.getField()
    Term.moveCursor(29, 26)
    Term.write("[")
    Term.write(Str)
    Term.write("]")
    Term.print(25, 27, Str)
end
-- quit()
