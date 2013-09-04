package main
import(
        "os"
)
func main(){
        //Configuration and proper error handlign missing
        link,err:=NewLink()
        if err!=nil{
                print( "Oh oh")
        }
        cli,err:=NewCli("dp2","[DP2]",*link)
        if err!=nil{
                print( "Oh oh")
        }
        scripts,err:=link.Scripts()
        if err!=nil{
                print( "Oh oh")
        }
        cli.AddScripts(scripts)
        cli.Run(os.Args[1:])
}

