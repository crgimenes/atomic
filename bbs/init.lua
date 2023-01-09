Term = require("term")
require "sysop_area"

if (getEnv("LANG") == "") then
    -- Term.setOutputMode("CP850")
    Term.setOutputMode("CP437")
    -- Term.setOutputMode("UTF-8") -- default
end

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

function MainMenu()
    clearTriggers()
    trigger("1", Runiptclient)
    trigger("2", SysopArea)
    trigger("3", ExitConnection)
    Term.write("\27[37;40m")
    Term.cls()

    Term.print(5, 8, "1 show shared terminal")
    Term.print(6, 8, "2 sysop area")
    Term.print(7, 8, "3 quit")

    Term.write("\27[35;40m")
    Term.print(15, 8, "option: ")
    Term.write("\27[37;40m")

end

function ExitConnection()
    local u = getUser()
    logf("quit user %s", u.nickname)
    Term.cls()
    Term.write("\r\nbye!\r\n")
    quit()
end

function Runiptclient()
    execWithTriggers("iptclient")
    MainMenu()
end

MainMenu()
