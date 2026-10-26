import {api,state} from './browser.mjs';
// Files were intentionally regenerated. Ask real application services to
// discard only their stale derived records in this isolated documentation DB.
console.log('fingerprints reset',await api('/videoFingerprint/resetAll',undefined,'DELETE'));
const tasks=await api('/videoTranscode/list');
if(tasks.length)console.log('old demo tasks removed',await api('/videoTranscode/deleteBatch',{ids:tasks.map(x=>x.id)}));
console.log('photo tasks added',await api('/videoTranscode/add',{resourceIds:state.resources.slice(0,12).map(x=>x.id)}));
