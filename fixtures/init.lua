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


trigger("q", exitConnection)
trigger("a", testPrint)

-- quit()
