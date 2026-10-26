import fs from 'node:fs';
import path from 'node:path';
import {api,state,run} from './browser.mjs';
const report={date:'2026-10-26',checks:[],limitations:['远程刮削站点、外部 AI 服务、TVBox 设备、GPU 与 Wails 原生窗口未做端到端验证。']};
const check=(name,data)=>{report.checks.push({name,data});console.log(name,JSON.stringify(data));};
const lib=state.libraries[0].id;
check('metadata-start',await api('/videoMetadata/run',{scopeMode:'selected',filesBasesIds:[lib],runMode:'all',maxItemsPerRun:100}));
for(let n=0;n<50;n++){const s=await api('/videoMetadata/taskStatus');if(s.status!=='running'){check('metadata-result',s);break;}await new Promise(r=>setTimeout(r,500));}
check('metadata-stats',await api('/videoMetadata/stats'));
check('fingerprint-rescan',await api('/videoFingerprint/reScan?files_bases_id='+lib));
check('fingerprint-compute',await api('/videoFingerprint/triggerCompute?batch_size=100&files_bases_id='+lib));
for(let n=0;n<240;n++){const s=await api('/videoFingerprint/taskStatus');if(!s.running){check('fingerprint-result',s);break;}await new Promise(r=>setTimeout(r,500));}
check('duplicate-groups',await api('/videoFingerprint/queryDuplicates?files_bases_id='+lib+'&match_mode=loose&threshold=15&duration_first=true&page=1&limit=10'));
const tasks=await api('/videoTranscode/list');const draft=tasks.find(x=>x.status==='draft');
if(draft){
 check('transcode-start',await api('/videoTranscode/start',{ids:[draft.id],enableGpu:false}));
 for(let n=0;n<100;n++){const task=(await api('/videoTranscode/list')).find(x=>x.id===draft.id);if(!['draft','queued','probing','transcoding','verifying','replacing'].includes(task.status)){check('transcode-result',task);break;}await new Promise(r=>setTimeout(r,500));}
}
check('preview-ready',await api('/videoPreviews/'+state.resources[0].dramaSeries[0].id));
check('single-tag-persisted',(await api('/resource/info/'+state.resources[0].id)).tags.map(x=>x.name));
check('batch-tags-persisted',await Promise.all(state.resources.slice(0,3).map(async r=>({title:r.title,tags:(await api('/resource/info/'+r.id)).tags.map(t=>t.name)}))));
fs.writeFileSync(path.join(run,'verification.json'),JSON.stringify(report,null,2));
