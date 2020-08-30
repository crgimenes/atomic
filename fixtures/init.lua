print("print in server terminal")
write("\r\nthis is a test write to client instance\r\n")

function exitConnection()
    print("quit user")
    write("\r\nbye!\r\n")
    quit()
end

function testPrint()
    print("testPrint")
    write("\r\ntest print\r\n")
end

echo = false
function toggleEcho() 
    echo = not echo
    setEcho(echo)
end

write("\r\n")
write("[e] toggle echo on/off\r\n")
write("[a] print test string\r\n")
write("[q] quit\r\n")
write("choose an option\r\n")


trigger("e", toggleEcho)
trigger("q", exitConnection)
trigger("a", testPrint)

-- quit()
