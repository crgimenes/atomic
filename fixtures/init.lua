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
write("\r\n")


-- writeFromASCII("nonfree/squiddy.ans")

Str = ""

function ExitConnection()
    print("quit user")
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
    write("\0277\27[0;0H")
    write(os.date('%Y-%m-%d %H:%M:%S UTC'))
    -- os.date("%Y-%m-%dT%H:%m:%S.000 %z"
    write("\0278")
    -- rmTrigger("clock")
end

write("[1] toggle echo on/off\r\n")
write("[2] print test string\r\n")
write("[3] quit\r\n")
write("choose an option\r\n")


timer("clock", 500, Clock)




trigger("1", ToggleEcho)
trigger("2", TestPrint)
trigger("3", ExitConnection)

write("\nenter a string:")
Str = getField()
write("\n\n\n\r\n")
write("[")
write(Str)
write("]")
write("\n\n\n\r\n")
-- quit()
