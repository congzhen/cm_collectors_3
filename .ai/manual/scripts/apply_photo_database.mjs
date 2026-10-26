import fs from 'node:fs';
import path from 'node:path';
import {api,state,base,run} from './browser.mjs';
const manifest=JSON.parse(fs.readFileSync(path.join(base,'photo-sources.json'),'utf8'));
const scenes=manifest.sources.filter(x=>x.kind==='scene').sort((a,b)=>a.index-b.index);
const portraits=manifest.sources.filter(x=>x.kind==='portrait').sort((a,b)=>a.index-b.index);
const image64=file=>'data:image/jpeg;base64,'+fs.readFileSync(path.join(base,file)).toString('base64');
for(let n=0;n<state.resources.length;n++){
 const r=await api('/resource/info/'+state.resources[n].id);const image=scenes[n%24];
 await api('/resource/update',{resource:{...r,subtitle:'摄影素材演示 · 独立手册数据',coverPosterWidth:600,coverPosterHeight:850,abstract:`《${r.title}》为使用手册虚构示例。封面采用授权摄影，展示资源整理、标签筛选与播放操作，不代表真实影片或与照片人物有关的作品。`},photoBase64:image64(image.poster_file),performers:r.performers.map(x=>x.id),directors:r.directors.map(x=>x.id),tags:r.tags.map(x=>x.id),dramaSeries:r.dramaSeries.map(x=>({id:x.id,src:x.src}))},'PUT');
}
for(let n=0;n<state.people.length;n++){
 const p=await api('/performer/info/'+state.people[n].id);
 await api('/performer/update',{performer:{...p,introduction:'虚构演示资料；头像为授权摄影，仅作界面示例，不表示照片人物的真实姓名、国籍、职业或关联作品。'},photoBase64:image64(portraits[n].local_file),tagIds:(p.tags||[]).map(x=>x.id)},'PUT');
}
// These are only stale derived previews of this generated demo. Rebuild them
// with the real application queue so screenshots cannot retain the old art.
for(const r of state.resources.slice(0,16)){
 await api('/videoPreviews/'+r.dramaSeries[0].id,undefined,'DELETE');
 await api('/videoPreviews/'+r.dramaSeries[0].id,{interval:10});
}
fs.writeFileSync(path.join(run,'photography-applied.json'),JSON.stringify({date:'2026-10-26',resources:96,people:24,sources:'photo-sources.json'},null,2));
console.log('96 covers and 24 portraits updated; original resource IDs, tags, people and moments preserved.');
