package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nomos/nomos/internal/fsx"
	"github.com/nomos/nomos/internal/model"
	"github.com/spf13/cobra"
)

var version = "dev"
var commit = "none"
var date = "unknown"

func Execute() { _ = newRoot().Execute() }
func newRoot() *cobra.Command {
	root := &cobra.Command{Use: "nomos", Short: "Nomos Cosmos CLI", Long: "Nomos verwaltet lokale Cosmos Repositories."}
	root.AddCommand(versionCmd(), cosmosCmd(), domainCmd(), serviceCmd(), validateCmd(), graphCmd(), verifyCmd(), serveCmd())
	return root
}
func versionCmd() *cobra.Command { return &cobra.Command{Use: "version", Run: func(cmd *cobra.Command, args []string) { fmt.Printf("version=%s\ncommit=%s\ndate=%s\ngo=%s\nosarch=%s/%s\n", version, commit, date, runtime.Version(), runtime.GOOS, runtime.GOARCH) }} }

func cosmosCmd() *cobra.Command { c := &cobra.Command{Use: "cosmos"}; var force, git bool
	init := &cobra.Command{Use: "init <path>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error { p:=args[0]; if st,err:=os.Stat(p); err==nil&&st.IsDir(){entries,_:=os.ReadDir(p); if len(entries)>0&&!force{return fmt.Errorf("Zielpfad ist nicht leer")}}; os.MkdirAll(filepath.Join(p,"domains"),0o755); os.MkdirAll(filepath.Join(p,".nomos/cache"),0o755); os.MkdirAll(filepath.Join(p,".nomos/index"),0o755); os.MkdirAll(filepath.Join(p,".nomos/evidence"),0o755); co:=model.Cosmos{ID:"cosmos-local",Type:"cosmos",Name:"Local Cosmos",Version:"0.1.0",Status:"draft",Owner:"unknown",Summary:"Lokaler Nomos Cosmos.",Domains:[]string{}}; _=fsx.WriteYAML(filepath.Join(p,"cosmos.yaml"),co); _=os.WriteFile(filepath.Join(p,"README.md"),[]byte("# Cosmos\n"),0o644); if git { _=os.WriteFile(filepath.Join(p,".gitkeep"),[]byte{},0o644)}; fmt.Println("Cosmos wurde erstellt:",p); return nil }}
	init.Flags().BoolVar(&force,"force",false,""); init.Flags().BoolVar(&git,"git",false,""); c.AddCommand(init)
	info:=&cobra.Command{Use:"info",RunE: func(cmd *cobra.Command,args []string) error { p,_:=cmd.Flags().GetString("path"); if p==""{p="."}; var co model.Cosmos; if err:=fsx.ReadYAML(filepath.Join(p,"cosmos.yaml"),&co);err!=nil{return err}; fmt.Printf("id=%s name=%s version=%s status=%s owner=%s domains=%d\n",co.ID,co.Name,co.Version,co.Status,co.Owner,len(co.Domains)); return nil }}; info.Flags().String("path",".",""); c.AddCommand(info)
	doc:=&cobra.Command{Use:"doctor",RunE: func(cmd *cobra.Command,args []string) error { p,_:=cmd.Flags().GetString("path"); if p==""{p="."}; if _,e:=os.Stat(filepath.Join(p,"cosmos.yaml"));e!=nil{fmt.Println("ERROR cosmos.yaml fehlt"); os.Exit(1)}; fmt.Println("OK cosmos.yaml gefunden"); if _,e:=os.Stat(filepath.Join(p,".git"));e!=nil{fmt.Println("WARNING Git Repository nicht initialisiert")} else {fmt.Println("OK Git Repository gefunden")}; return nil }}; doc.Flags().String("path",".",""); c.AddCommand(doc)
	return c }

func domainCmd() *cobra.Command { c:=&cobra.Command{Use:"domain"}; var force bool
	add:=&cobra.Command{Use:"add <dns>",Args:cobra.ExactArgs(1),RunE: func(cmd *cobra.Command,args []string) error { dns:=args[0]; if !strings.Contains(dns,"."){return fmt.Errorf("ungueltiger DNS Name")}; p,_:=cmd.Flags().GetString("path"); owner,_:=cmd.Flags().GetString("owner"); ddir:=filepath.Join(p,"domains",dns); if _,e:=os.Stat(ddir);e==nil&&!force{return fmt.Errorf("Domain existiert bereits")}; os.MkdirAll(filepath.Join(ddir,"services"),0o755); d:=model.Domain{ID:"domain-"+strings.ReplaceAll(dns,".","-"),Type:"domain",Name:dns,Version:"0.1.0",Status:"draft",Owner:owner,DNSName:dns,Summary:"Nomos Domaene "+dns+"."}; _=fsx.WriteYAML(filepath.Join(ddir,"domain.yaml"),d); _=os.WriteFile(filepath.Join(ddir,"README.md"),[]byte("# Domain\n"),0o644); return nil }}
	add.Flags().String("path",".",""); add.Flags().String("owner","unknown",""); add.Flags().BoolVar(&force,"force",false,""); c.AddCommand(add)
	c.AddCommand(&cobra.Command{Use:"list",RunE: func(cmd *cobra.Command,args []string) error { p,_:=cmd.Flags().GetString("path"); ents,_:=os.ReadDir(filepath.Join(p,"domains")); for _,e:= range ents { if e.IsDir(){fmt.Println(e.Name())}}; return nil }})
	return c }
func serviceCmd() *cobra.Command { c:=&cobra.Command{Use:"service"}; var force bool
	add:=&cobra.Command{Use:"add <name>",Args:cobra.ExactArgs(1),RunE: func(cmd *cobra.Command,args []string) error { name:=args[0]; dom,_:=cmd.Flags().GetString("domain"); p,_:=cmd.Flags().GetString("path"); owner,_:=cmd.Flags().GetString("owner"); sdir:=filepath.Join(p,"domains",dom,"services",name); if _,e:=os.Stat(sdir);e==nil&&!force{return fmt.Errorf("Service existiert bereits")}; for _,d:= range []string{"capabilities","requirements","rules","processes","skills","findings","evidence"}{os.MkdirAll(filepath.Join(sdir,d),0o755)}; s:=model.Service{ID:"service-"+name,Type:"service",Name:name,Version:"0.1.0",Status:"draft",Owner:owner,Summary:"Nomos Service "+name+"."}; _=fsx.WriteYAML(filepath.Join(sdir,"service.yaml"),s); _=os.WriteFile(filepath.Join(sdir,"README.md"),[]byte("# Service\n"),0o644); return nil }}
	add.Flags().String("domain","",""); add.Flags().String("path",".",""); add.Flags().String("owner","unknown",""); add.Flags().BoolVar(&force,"force",false,""); _=add.MarkFlagRequired("domain"); c.AddCommand(add); return c }

func validateCmd() *cobra.Command { return &cobra.Command{Use:"validate",RunE: func(cmd *cobra.Command,args []string) error { p,_:=cmd.Flags().GetString("path"); if p==""{p="."}; findings:=[]model.Finding{}; if _,e:=os.Stat(filepath.Join(p,"cosmos.yaml"));e!=nil{findings=append(findings, model.Finding{Severity:"error",Code:"COSMOS_MISSING",Message:"cosmos.yaml fehlt",Path:"cosmos.yaml",Recommendation:"Erzeuge eine Cosmos Struktur."})}; outFmt,_:=cmd.Flags().GetString("format"); if outFmt=="json"{b,_:=json.MarshalIndent(map[string]any{"status":"ok","findings":findings},"","  "); fmt.Println(string(b))} else {fmt.Println("Nomos Validierung"); for _,f:= range findings {fmt.Printf("%s %s\n",strings.ToUpper(f.Severity),f.Message)}}; if len(findings)>0{os.Exit(1)}; return nil } } }
func graphCmd() *cobra.Command { return &cobra.Command{Use:"graph",RunE: func(cmd *cobra.Command,args []string) error { p,_:=cmd.Flags().GetString("path"); fmt.Printf("graph TD\n  cosmos[\"Cosmos: %s\"]\n",p); return nil }} }
func verifyCmd() *cobra.Command { v:=&cobra.Command{Use:"verify"}; d:=&cobra.Command{Use:"domain <dns>",Args:cobra.ExactArgs(1),RunE: func(cmd *cobra.Command,args []string) error { dns:=args[0]; rec:="_nomos."+dns; txt,err:=net.DefaultResolver.LookupTXT(cmd.Context(),rec); p,_:=cmd.Flags().GetString("path"); os.MkdirAll(filepath.Join(p,".nomos/evidence"),0o755); status:="failed"; exp:="nomos-domain="+dns; for _,t:= range txt { if strings.Contains(t,exp){status="verified"} }; ev:=fmt.Sprintf("id: evidence-%s\ntype: evidence\nevidence_type: dns_verification\ndomain: %s\nrecord: %s\nstatus: %s\ntimestamp: \"%s\"\n",time.Now().UTC().Format("20060102-150405"),dns,rec,status,time.Now().UTC().Format(time.RFC3339)); _=os.WriteFile(filepath.Join(p,".nomos/evidence",strings.ReplaceAll(dns,".","-")+"-dns.yaml"),[]byte(ev),0o644); if err!=nil||status!="verified"{return fmt.Errorf("DNS Verifikation fehlgeschlagen")}; fmt.Println("DNS Verifikation erfolgreich"); return nil }}; d.Flags().String("path",".",""); v.AddCommand(d); return v }
func serveCmd() *cobra.Command { c:=&cobra.Command{Use:"serve",RunE: func(cmd *cobra.Command,args []string) error { listen,_:=cmd.Flags().GetString("listen"); mux:=http.NewServeMux(); mux.HandleFunc("/health", func(w http.ResponseWriter,r *http.Request){_ = json.NewEncoder(w).Encode(map[string]string{"status":"ok","service":"nomos","version":version})}); mux.HandleFunc("/api/v1/validate", func(w http.ResponseWriter,r *http.Request){_ = json.NewEncoder(w).Encode(map[string]string{"status":"ok"})}); fmt.Println("Server lauscht auf",listen); return http.ListenAndServe(listen,mux) }}; c.Flags().String("listen","127.0.0.1:8080",""); return c }
