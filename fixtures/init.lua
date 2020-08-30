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



trigger("e", toggleEcho)
trigger("q", exitConnection)
trigger("a", testPrint)

-- quit()
