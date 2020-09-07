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

str = ""

function exitConnection()
    print("quit user")
    write("\r\nbye!\r\n")
    quit()
end

function testPrint()
    print("testPrint")
    write("\r\ntest print àáéíóúü~ãõç\r\n")
    write("str: ")
    write(str)
    write("\r\n")
end

echo = false
function toggleEcho() 
    echo = not echo
    setEcho(echo)
end

setANSI(1,4,31)
write("[1] toggle echo on/off\r\n")
write("[2] print test string\r\n")
write("[3] quit\r\n")
write("choose an option\r\n")


trigger("1", toggleEcho)
trigger("2", testPrint)
trigger("3", exitConnection)

write("enter a string:")
str = getField()
write("\n\n\n\r\n")
write(str)
write("\n\n\n\r\n")
-- quit()
