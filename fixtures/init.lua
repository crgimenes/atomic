cls()
print("print in server terminal")
write("\r\nthis is a test write to client instance\r\n")
write(" ██████╗██████╗  ██████╗    ███████╗████████╗██╗   ██████╗ ██████╗\r\n")
write("██╔════╝██╔══██╗██╔════╝    ██╔════╝╚══██╔══╝██║   ██╔══██╗██╔══██╗\r\n")
write("██║     ██████╔╝██║  ███╗   █████╗     ██║   ██║   ██████╔╝██████╔╝\r\n")
write("██║     ██╔══██╗██║   ██║   ██╔══╝     ██║   ██║   ██╔══██╗██╔══██╗\r\n")
write("╚██████╗██║  ██║╚██████╔╝██╗███████╗   ██║   ██║██╗██████╔╝██║  ██║\r\n")
write(" ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝╚══════╝   ╚═╝   ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝\r\n")
write("crg@crg.eti.br @crgimenes\r\n")
write("██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██  ██\r\n")


writeFromASCII("nonfree/squiddy.ans")

Str = ""

function ExitConnection()
    print("quit user")
    write("\r\nbye!\r\n")
    quit()
end

function TestPrint()
    print("testPrint")
    write("\r\ntest print àáéíóúü~ãõç\r\n")
    write("Str: ")
    write(Str)
    write("\r\n")
end

Echo = false
function ToggleEcho()
    Echo = not Echo
    setEcho(Echo)
end

setANSI(1,4,31)
write("[1] toggle echo on/off\r\n")
write("[2] print test string\r\n")
write("[3] quit\r\n")
write("choose an option\r\n")


trigger("1", ToggleEcho)
trigger("2", TestPrint)
trigger("3", ExitConnection)

write("enter a string:")
Str = getField()
write("\n\n\n\r\n")
write(Str)
write("\n\n\n\r\n")
-- quit()
