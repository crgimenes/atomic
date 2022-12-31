local function run_ipt_client()
    exec("iptclient")
end

local function back()
    Menu()
end

function Main_menu()
    clearTriggers()
    Term.cls()
    Term.write("\r\nmain menu\r\n")
    Term.write("[1] love coding\r\n")
    Term.write("[0] back\r\n")


    trigger("1", run_ipt_client)
    trigger("0", back)

end
