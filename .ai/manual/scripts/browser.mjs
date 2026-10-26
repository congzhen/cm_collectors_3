import fs from 'node:fs';
import path from 'node:path';
import {createRequire} from 'node:module';
import {fileURLToPath} from 'node:url';
export const base=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
export const run=path.join(base,'.local/runtime');
export const state=JSON.parse(fs.readFileSync(path.join(run,'demo-state.json'),'utf8'));
const require=createRequire(path.resolve(base,'../../cm_collectors_html/package.json'));
const {chromium}=require('@playwright/test');
export async function api(route,data,method=data===undefined?'GET':'POST') {
 const r=await fetch('http://127.0.0.1:18345/api'+route,{method,headers:{'Content-Type':'application/json'},body:data===undefined?undefined:JSON.stringify(data),signal:AbortSignal.timeout(60000)});
 const j=await r.json();if(!j.status)throw Error(route+':'+j.msg);return j.data;
}
export async function browser() {
 const exe=[chromium.executablePath(),'C:/Program Files/Google/Chrome/Application/chrome.exe','C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe'].find(p=>fs.existsSync(p));
 return chromium.launch({headless:true,executablePath:exe,args:['--no-proxy-server']});
}
export async function pageReady(p){await p.locator('.el-loading-mask:visible').first().waitFor({state:'hidden',timeout:15000}).catch(()=>{});await p.waitForTimeout(900);}
export async function home(p){await p.goto('http://127.0.0.1:18346/',{waitUntil:'networkidle'}); await p.getByRole('tab',{name:'山海影像',exact:true}).click();await pageReady(p);}
export async function shot(p,name){fs.mkdirSync(path.join(base,'assets/screenshots'),{recursive:true});await pageReady(p);await p.screenshot({path:path.join(base,'assets/screenshots',name+'.jpg'),type:'jpeg',quality:86,animations:'disabled'});console.log('screenshot: '+name);}
if(process.argv[1]===fileURLToPath(import.meta.url)) {
 const b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});
 p.on('pageerror',e=>console.log('PAGE ERROR',e.message));await home(p);await shot(p,'home');console.log((await p.locator('body').innerText()).slice(0,6000));await b.close();
}
