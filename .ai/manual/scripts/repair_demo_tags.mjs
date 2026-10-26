import {api,state} from './browser.mjs';
const names=[['自然','旅行','城市','人文','建筑','生活'],['山川','海岸','森林','湖泊','夜景','街巷'],['待看','已看','精选','周末片单','学习参考','收藏']];
for(let lib=0;lib<4;lib++){
 const tags=await api('/tag/list/filesBasesId/'+state.libraries[lib].id);
 for(let i=0;i<24;i++)await api('/resource/update/tag',{resourceId:state.resources[lib*24+i].id,tags:names.map((n,g)=>tags.find(t=>t.name===n[(i*(g===1?5:1))%6]).id)},'PUT');
}
console.log('96 resources have tags across all three categories');
