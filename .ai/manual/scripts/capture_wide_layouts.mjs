// Real API/UI capture using saved landscape photography. Restores demo covers
// and library settings afterwards; never points at the development service.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';
import {api,browser,home,shot,base,run,state} from './browser.mjs';
const backupFile=path.join(run,'wide-layouts-backup.json');
const receiptFile=path.join(run,'wide-layouts-verification.json');
const lib=state.libraries[0].id;
const modes=['coverPosterCinemaGallery','coverPosterBoxWideSeparate','coverPosterSimpleExpand'];
const sources=JSON.parse(fs.readFileSync(path.join(base,'photo-sources.json'),'utf8')).sources.filter(x=>x.kind==='scene').sort((a,b)=>a.index-b.index);
const hash=buffer=>crypto.createHash('sha256').update(buffer).digest('hex');
async function coverBytes(r){
 const response=await fetch(`http://127.0.0.1:18345/api/resCoverPoster/${encodeURIComponent(r.filesBases_id)}/${encodeURIComponent(r.coverPoster)}`);
 if(!response.ok)throw Error('Cannot back up cover '+r.id);
 return Buffer.from(await response.arrayBuffer());
}
async function update(r,bytes,width,height){
 await api('/resource/update',{resource:{...r,coverPosterWidth:width,coverPosterHeight:height},photoBase64:'data:image/jpeg;base64,'+bytes.toString('base64'),performers:r.performers.map(x=>x.id),directors:r.directors.map(x=>x.id),tags:r.tags.map(x=>x.id),dramaSeries:r.dramaSeries.map(x=>({id:x.id,src:x.src}))},'PUT');
}
async function restore(saved){
 for(const entry of saved.resources){
  await update(entry.resource,Buffer.from(entry.cover,'base64'),entry.resource.coverPosterWidth,entry.resource.coverPosterHeight);
  const current=await api('/resource/info/'+entry.resource.id);
  if(hash(await coverBytes(current))!==entry.sha256)throw Error('Cover restore mismatch '+current.id);
 }
 await api('/filesBases/setConfig/filesBases',{id:lib,config:saved.config},'PUT');
 saved.restored=true;fs.writeFileSync(backupFile,JSON.stringify(saved));
}
if(process.argv.includes('--restore')){
 await restore(JSON.parse(fs.readFileSync(backupFile,'utf8')));console.log('Saved demo state restored.');
}else{
 if(fs.existsSync(backupFile)&&!JSON.parse(fs.readFileSync(backupFile,'utf8')).restored)throw Error('Incomplete previous capture; run with --restore first.');
 const config=await api('/filesBases/config/'+lib+'/filesBases');
 const saved={config,resources:[],restored:false};
 const selected=state.resources.filter(r=>r.filesBases_id===lib);
 if(selected.length<12)throw Error('Need at least a full screen of demo resources.');
 for(const item of selected){const r=await api('/resource/info/'+item.id);const bytes=await coverBytes(r);saved.resources.push({resource:r,cover:bytes.toString('base64'),sha256:hash(bytes)});}
 fs.writeFileSync(backupFile,JSON.stringify(saved));
 const receipt={library:lib,source_manifest:'photo-sources.json',resources:selected.length,display_cover:{width:300,height:169,fit:'cover'},modes:[],restored:false};
 let b;
 try{
  for(let n=0;n<saved.resources.length;n++){
   const entry=saved.resources[n];const source=sources[n%sources.length];const bytes=fs.readFileSync(path.join(base,source.local_file));
   if(hash(bytes)!==source.sha256)throw Error('Source checksum mismatch');
   await update(entry.resource,bytes,960,640);
  }
  b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});
  const errors=[];p.on('pageerror',e=>errors.push(e.message));
  const original=JSON.parse(config||'{}');
  for(const mode of modes){
   await api('/filesBases/setConfig/filesBases',{id:lib,config:JSON.stringify({...original,resourcesShowMode:mode,coverPosterWidthStatus:true,coverPosterWidthBase:300,coverPosterHeightStatus:true,coverPosterHeightBase:169,coverImageFit:'cover',pageLimit:24})},'PUT');
   await home(p);
   await p.locator('.content-right-click-menu img').first().waitFor({state:'visible'});
   await p.waitForFunction(()=>[...document.querySelectorAll('.content-right-click-menu img')].filter(img=>img.complete&&img.naturalWidth>img.naturalHeight).length>=9);
   const images=await p.locator('.content-right-click-menu img').evaluateAll(imgs=>imgs.filter(img=>img.complete&&img.naturalWidth>0).map(img=>({natural:[img.naturalWidth,img.naturalHeight],display:[Math.round(img.width),Math.round(img.height)]})));
   if(images.length<12||images.some(img=>img.natural[0]<=img.natural[1]))throw Error('Landscape covers not ready');
   await shot(p,'layout-'+mode);receipt.modes.push({mode,images});
  }
  if(errors.length)throw Error(JSON.stringify(errors));
 }finally{
  if(b)await b.close();await restore(saved);receipt.restored=true;receipt.completed_at=new Date().toISOString();fs.writeFileSync(receiptFile,JSON.stringify(receipt,null,2));
 }
 console.log(JSON.stringify({captured:receipt.modes.map(x=>x.mode),resources:selected.length,restored:receipt.restored}));
}
