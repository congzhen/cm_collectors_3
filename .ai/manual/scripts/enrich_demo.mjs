import fs from 'node:fs';
import path from 'node:path';
import {api, state, run} from './browser.mjs';
const done=path.join(run,'enriched.json');
if(fs.existsSync(done)){console.log('Enriched demo reused');process.exit(0);}
const app=await api('/app/data');
for(const x of app.filesBases.filter(x=>x.name==='Default'))await api('/filesBases/delete/'+x.id,undefined,'DELETE');
await api('/filesBases/sort',{sortData:state.libraries.map((x,i)=>({id:x.id,sort:i}))},'PUT');
const config=await api('/app/getConfig');config.playCloudMode='mp4';config.appearanceStyle='modern';await api('/app/setConfig',config,'PUT');
for(let i=0;i<16;i++){
 const r=state.resources[i];
 await api('/videoMoments',{resourceId:r.id,dramaSeriesId:r.dramaSeries[0].id,title:`${r.title} · ${i%2?'光影片段':'开场时刻'}`,note:'手册演示收藏：记录构图与配色。',start:i%2?8:3,end:i%2?18:0});
 await api('/videoPreviews/'+r.dramaSeries[0].id,{interval:10});
}
console.log('Moments and previews created');
console.log(await api('/videoTranscode/add',{resourceIds:state.resources.slice(0,12).map(x=>x.id)}));
for(const r of state.resources.slice(0,12))await api('/tvbox/recommend/add/'+r.id,{});
const existing=await api('/cronJobs/list');
if(!existing.length)for(const [i,type]of ['import','scraperResource','scraperPerformer','clear','videoFingerprint','aiTag','videoMetadata','clearPerformerAvatarCache'].entries()){
 try{await api('/cronJobs/create',{filesBases_id:state.libraries[0].id,filesBasesIds:[state.libraries[0].id],scopeMode:'selected',jobs_type:type,cron_expression:`0 ${i*5} 3 * * *`,configJsonData:type==='videoMetadata'?JSON.stringify({maxItemsPerRun:20,runMode:'missing_stale'}):'',status:false});}catch(e){console.log(e.message);}
}
const result={moments:16,transcodeDrafts:12,tvbox:12,cron:(await api('/cronJobs/list')).length};
fs.writeFileSync(done,JSON.stringify(result,null,2));console.log(result);
