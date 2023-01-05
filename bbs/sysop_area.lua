local function run_ipt_client()
    exec("iptclient")
end

local function back()
    MainMenu()
end

function SysopMenu()
    clearTriggers()
    Term.cls()
    trigger("1", run_ipt_client)
    trigger("2", run_test)
    trigger("0", back)
    Term.write("\r\nmain menu\r\n")
    Term.write("[1] live coding\r\n")
    Term.write("[2] run test\r\n")
    Term.write("[0] back\r\n")
end

function run_test()
    SysopMenu()
    exec("ls")
end

function SysopArea()
    if hasGroup("sysop") then
        SysopMenu()
        return
    end
    MainMenu()
    Term.write("\r\nyou are not a sysop\r\n")
end
