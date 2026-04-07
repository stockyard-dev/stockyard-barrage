package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Barrage</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{padding:1.5rem;max-width:960px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center}
.search{flex:1;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.test{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.5rem;transition:border-color .2s}
.test:hover{border-color:var(--leather)}
.test-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.test-name{font-size:.85rem;font-weight:700}
.test-url{font-size:.65rem;color:var(--cd);margin-top:.1rem}
.test-config{font-size:.55rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.6rem;flex-wrap:wrap}
.test-actions{display:flex;gap:.3rem;flex-shrink:0}
.method-badge{font-size:.5rem;padding:.12rem .35rem;text-transform:uppercase;letter-spacing:1px;border:1px solid;font-weight:700}
.method-badge.GET{border-color:var(--green);color:var(--green)}.method-badge.POST{border-color:var(--blue);color:var(--blue)}.method-badge.PUT{border-color:var(--gold);color:var(--gold)}.method-badge.DELETE{border-color:var(--red);color:var(--red)}
.run{background:var(--bg);border:1px solid var(--bg3);padding:.5rem .7rem;margin-top:.3rem;font-size:.6rem}
.run-stats{display:flex;gap:.6rem;flex-wrap:wrap;margin-top:.2rem}
.run-stat{display:flex;flex-direction:column;align-items:center}.run-stat-v{font-weight:700;font-size:.7rem}.run-stat-l{font-size:.45rem;color:var(--cm);text-transform:uppercase}
.badge{font-size:.5rem;padding:.12rem .35rem;text-transform:uppercase;letter-spacing:1px;border:1px solid}
.badge.running{border-color:var(--gold);color:var(--gold)}.badge.done{border-color:var(--green);color:var(--green)}.badge.failed{border-color:var(--red);color:var(--red)}
.btn{font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}
.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.btn-run{border-color:var(--green);color:var(--green)}.btn-run:hover{background:var(--green);color:#fff}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:500px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.row3{display:grid;grid-template-columns:1fr 1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
@media(max-width:600px){.stats{grid-template-columns:repeat(2,1fr)}.row2,.row3{grid-template-columns:1fr}}
</style></head><body>
<div class="hdr"><h1><span>&#9670;</span> BARRAGE</h1><button class="btn btn-p" onclick="openForm()">+ New Test</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar"><input class="search" id="search" placeholder="Search tests..." oninput="render()"></div>
<div id="tests"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()"><div class="modal" id="mdl"></div></div>
<script>
var A='/api',tests=[],runs={},editId=null;

async function load(){
var r=await fetch(A+'/tests').then(function(r){return r.json()});tests=r.tests||[];
for(var i=0;i<tests.length;i++){
var rr=await fetch(A+'/tests/'+tests[i].id+'/runs?limit=3').then(function(r){return r.json()}).catch(function(){return{runs:[]}});
runs[tests[i].id]=rr.runs||[];
}
renderStats();render();
}

function renderStats(){
var total=tests.length;
var totalRuns=0;tests.forEach(function(t){totalRuns+=t.run_count||0});
var totalReqs=0;Object.values(runs).forEach(function(rs){rs.forEach(function(r){totalReqs+=r.total_requests||0})});
document.getElementById('stats').innerHTML=[
{l:'Tests',v:total},{l:'Total Runs',v:totalRuns},{l:'Requests Sent',v:totalReqs>999?(totalReqs/1000).toFixed(1)+'k':totalReqs}
].map(function(x){return '<div class="st"><div class="st-v">'+x.v+'</div><div class="st-l">'+x.l+'</div></div>'}).join('');
}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var f=tests;
if(q)f=f.filter(function(t){return(t.name||'').toLowerCase().includes(q)||(t.url||'').toLowerCase().includes(q)});
if(!f.length){document.getElementById('tests').innerHTML='<div class="empty">No load tests. Create one to get started.</div>';return;}
var h='';f.forEach(function(t){
h+='<div class="test"><div class="test-top"><div style="flex:1">';
h+='<div class="test-name">'+esc(t.name)+'</div>';
h+='<div class="test-url"><span class="method-badge '+t.method+'">'+t.method+'</span> '+esc(t.url)+'</div>';
h+='</div><div class="test-actions">';
h+='<button class="btn btn-sm btn-run" onclick="runTest(''+t.id+'')">&#9654; Run</button>';
h+='<button class="btn btn-sm" onclick="openEdit(''+t.id+'')">Edit</button>';
h+='<button class="btn btn-sm" onclick="del(''+t.id+'')" style="color:var(--red)">&#10005;</button>';
h+='</div></div>';
h+='<div class="test-config">';
h+='<span>'+t.concurrency+' concurrent</span>';
h+='<span>'+t.requests+' requests</span>';
h+='<span>'+t.run_count+' runs</span>';
h+='</div>';
var tRuns=runs[t.id]||[];
tRuns.forEach(function(r){
h+='<div class="run"><div style="display:flex;justify-content:space-between;align-items:center">';
h+='<span class="badge '+r.status+'">'+r.status+'</span>';
h+='<span style="color:var(--cm)">'+ft(r.started_at)+'</span></div>';
if(r.status==='done'){
h+='<div class="run-stats">';
h+='<div class="run-stat"><div class="run-stat-v" style="color:var(--green)">'+r.successes+'</div><div class="run-stat-l">OK</div></div>';
h+='<div class="run-stat"><div class="run-stat-v" style="color:var(--red)">'+r.failures+'</div><div class="run-stat-l">Fail</div></div>';
h+='<div class="run-stat"><div class="run-stat-v">'+r.avg_ms.toFixed(1)+'</div><div class="run-stat-l">Avg ms</div></div>';
h+='<div class="run-stat"><div class="run-stat-v">'+r.p99_ms.toFixed(1)+'</div><div class="run-stat-l">P99 ms</div></div>';
h+='<div class="run-stat"><div class="run-stat-v">'+r.req_per_sec.toFixed(0)+'</div><div class="run-stat-l">Req/s</div></div>';
h+='<div class="run-stat"><div class="run-stat-v">'+(r.duration_ms/1000).toFixed(1)+'s</div><div class="run-stat-l">Duration</div></div>';
h+='</div>';}
h+='</div>';
});
h+='</div>';
});
document.getElementById('tests').innerHTML=h;
}

async function runTest(id){await fetch(A+'/tests/'+id+'/run',{method:'POST'});setTimeout(load,500);}
async function del(id){if(!confirm('Delete this test and all runs?'))return;await fetch(A+'/tests/'+id,{method:'DELETE'});load();}

function formHTML(test){
var i=test||{name:'',url:'',method:'GET',headers:'',body:'',concurrency:10,requests:100};
var isEdit=!!test;
var h='<h2>'+(isEdit?'EDIT TEST':'NEW LOAD TEST')+'</h2>';
h+='<div class="fr"><label>Name *</label><input id="f-name" value="'+esc(i.name)+'" placeholder="e.g. API health check"></div>';
h+='<div class="row2"><div class="fr"><label>Method</label><select id="f-method">';
['GET','POST','PUT','DELETE','PATCH'].forEach(function(m){h+='<option value="'+m+'"'+(i.method===m?' selected':'')+'>'+m+'</option>';});
h+='</select></div><div class="fr"><label>URL *</label><input id="f-url" value="'+esc(i.url)+'" placeholder="https://api.example.com/health"></div></div>';
h+='<div class="row2"><div class="fr"><label>Concurrency</label><input id="f-conc" type="number" value="'+i.concurrency+'"></div>';
h+='<div class="fr"><label>Total Requests</label><input id="f-reqs" type="number" value="'+i.requests+'"></div></div>';
h+='<div class="fr"><label>Headers (JSON)</label><textarea id="f-headers" rows="2" placeholder='{"Authorization":"Bearer ..."}'>'+ esc(i.headers)+'</textarea></div>';
h+='<div class="fr"><label>Body</label><textarea id="f-body" rows="2" placeholder="Request body for POST/PUT">'+esc(i.body)+'</textarea></div>';
h+='<div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Create Test')+'</button></div>';
return h;
}

function openForm(){editId=null;document.getElementById('mdl').innerHTML=formHTML();document.getElementById('mbg').classList.add('open');document.getElementById('f-name').focus();}
function openEdit(id){var t=null;for(var j=0;j<tests.length;j++){if(tests[j].id===id){t=tests[j];break;}}if(!t)return;editId=id;document.getElementById('mdl').innerHTML=formHTML(t);document.getElementById('mbg').classList.add('open');}
function closeModal(){document.getElementById('mbg').classList.remove('open');editId=null;}

async function submit(){
var name=document.getElementById('f-name').value.trim();
var url=document.getElementById('f-url').value.trim();
if(!name||!url){alert('Name and URL are required');return;}
var body={name:name,url:url,method:document.getElementById('f-method').value,concurrency:parseInt(document.getElementById('f-conc').value)||10,requests:parseInt(document.getElementById('f-reqs').value)||100,headers:document.getElementById('f-headers').value.trim(),body:document.getElementById('f-body').value.trim()};
if(editId){await fetch(A+'/tests/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
else{await fetch(A+'/tests',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
closeModal();load();
}

function ft(t){if(!t)return'';try{var d=new Date(t);return d.toLocaleDateString('en-US',{month:'short',day:'numeric'})+' '+d.toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'})}catch(e){return t;}}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML;}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal();});
load();
</script><script>
(function(){
  fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
    if(!cfg||typeof cfg!=='object')return;
    if(cfg.dashboard_title){
      document.title=cfg.dashboard_title;
      var h1=document.querySelector('h1');
      if(h1){
        var inner=h1.innerHTML;
        var firstSpan=inner.match(/<span[^>]*>[^<]*<\/span>/);
        if(firstSpan){h1.innerHTML=firstSpan[0]+' '+cfg.dashboard_title}
        else{h1.textContent=cfg.dashboard_title}
      }
    }
  }).catch(function(){});
})();
</script>
</body></html>`
