local function run_ipt_client()
    exec("iptclient")
end

local function back()
    MainMenu()
end

function SysopMenu()
    clearTriggers()
    Term.cls()
    Term.write("\r\nmain menu\r\n")
    Term.write("[1] live coding\r\n")
    Term.write("[0] back\r\n")


    trigger("1", run_ipt_client)
    trigger("0", back)

end

function SysopArea()
    if hasGroup("sysop") then
        SysopMenu()
        return
    end
    MainMenu()
    Term.write("\r\nyou are not a sysop\r\n")
end
