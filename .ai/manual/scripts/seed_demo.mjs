import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const base=path.resolve(path.dirname(fileURLToPath(import.meta.url)),'..');
const run=path.join(base,'.local/runtime');
const origin='http://127.0.0.1:18345';
async function api(route, data, method=data===undefined?'GET':'POST') {
 const r=await fetch(origin+'/api'+route,{method,headers:{'Content-Type':'application/json'},body:data===undefined?undefined:JSON.stringify(data),signal:AbortSignal.timeout(20000)});
 const j=await r.json(); if(!r.ok || !j.status) throw new Error(route+': '+JSON.stringify(j));return j.data;
}
if(fs.existsSync(path.join(run,'demo-state.json'))) {console.log('Demo already seeded; reuse demo-state.json');process.exit(0);}
const app=await api('/app/data');
if(app.filesBases.some(x=>['山海影像','旅行图集','山海绘本','资料文件'].includes(x.name)))throw Error('Incomplete manual seed detected. Preserve the demo directory and inspect its state before retrying; no existing library will be overwritten.');
const perBase=app.performerBases[0].id;
await api('/performerBases/update',{id:perBase,name:'山海创作团队（虚构）',sort:0,status:true},'PUT');
const libraries=[];
for(const name of ['山海影像','旅行图集','山海绘本','资料文件']) libraries.push(await api('/filesBases/create',{name,mainPerformerBasesId:perBase,relatedPerformerBasesIds:[perBase]}));
const names=['林远','苏禾','陈星','江澄','沈舟','许岚','周野','顾宁','程夏','陆川','宋雨','方青','唐枫','白露','叶溪','徐森','季风','黎明','温晴','何月','夏泉','杜衡','秦海','莫言溪'];
const people=[];
for(let i=0;i<24;i++) people.push(await api('/performer/create',{performer:{performerBases_id:perBase,name:names[i],aliasName:`创作者${i+1}`,nationality:'China',careerPerformer:true,careerDirector:i%4===0,introduction:'虚构演示人物，用于展示演员与资源关联。',stars:3+i%3,status:true},photoBase64:'data:image/jpeg;base64,'+fs.readFileSync(path.join(run,`db/performerFace/manual-${String(i+1).padStart(2,'0')}.jpg`)).toString('base64'),tagIds:[]}));
const titles=JSON.parse(fs.readFileSync(path.join(run,'asset-manifest.json'),'utf8')).titles;
const resources=[];const allTags=[];
for(let lib=0;lib<libraries.length;lib++) {
 const id=libraries[lib].id;
 const groups={'题材':['自然','旅行','城市','人文','建筑','生活'],'场景':['山川','海岸','森林','湖泊','夜景','街巷'],'观看计划':['待看','已看','精选','周末片单','学习参考','收藏']};
 for(const [name,tags] of Object.entries(groups)) {
  await api('/tagClass/create',{filesBases_id:id,name,leftShow:true,sort:Object.keys(groups).indexOf(name),status:true});
  const cls=(await api('/tagClass/list/'+id)).find(x=>x.name===name);
  for(let i=0;i<tags.length;i++) await api('/tag/create',{tagClass_id:cls.id,name:tags[i],aiDescription:`画面或主题符合${tags[i]}`,aiEnabled:false,sort:i,status:true});
 }
 const tagList=await api('/tag/list/filesBasesId/'+id); allTags.push(tagList);
 for(let i=0;i<24;i++) {
  const mode=['movies','atlas','comic','files'][lib];
  const src=path.join(run,lib===0?`media/videos/CM-${String(i+1).padStart(3,'0')}.mp4`:lib===1?'media/atlas':lib===2?'media/comic':'media/README.txt');
  const tags=[groups['题材'][i%6],groups['场景'][(i*5)%6],groups['观看计划'][i%6]].map(name=>tagList.find(t=>t.name===name).id);
  const resource=await api('/resource/create',{resource:{filesBases_id:id,mode,title:titles[i]+(lib===1?' · 图集':lib===2?' · 绘本':lib===3?' · 资料':''),subtitle:'原创演示系列 · 完全独立的数据',issueNumber:`CM-${lib+1}${String(i+1).padStart(3,'0')}`,coverPosterMode:0,coverPosterWidth:480,coverPosterHeight:680,issuingDate:`202${3+i%4}-${String(i%12+1).padStart(2,'0')}-01`,country:['China','France','Iceland','Japan'][i%4],definition:['1080P','4K','720P'][i%3],stars:3+i%3,score:7+(i%20)/10,abstract:`《${titles[i]}》是一份独立生成的手册演示资源。通过山川、城市与旅行主题，演示资源整理、标签筛选、演员关联及播放操作。所有人物和作品均为虚构。`},photoBase64:'data:image/jpeg;base64,'+fs.readFileSync(path.join(run,`db/resCoverPoster/manual-${String(i+1).padStart(2,'0')}.jpg`)).toString('base64'),performers:[people[i].id,people[(i+5)%24].id],directors:[people[(i+4)%24].id],tags,dramaSeries:[{src}]});
  resources.push(resource);
 }
 await api('/filesBases/setConfig/filesBases',{id,config:JSON.stringify({resourcesShowMode:'coverPoster',pageLimit:48,leftColumnWidth:260,leftDisplay:['sort','diyTag','starRating','performer'],coverPosterHeightBase:230,coverPosterHeightStatus:true,coverPosterGap:12,coverPosterWidthBase:164,coverPosterData:[{name:'竖版海报',width:480,height:680,type:'default'},{name:'横版封面',width:960,height:640,type:'default'}],coverPosterDataDefaultSelect:0,coverDisplayTag:['tag'],coverDisplayTagAttribute:['definition','stars'],showVideoDuration:true,resourceDetailsShowMode:'popup',performerShowNum:12,videoPreviewGeneration:'manual',sampleStatus:false})},'PUT');
}
const state={origin,libraries,perBase,people,resources,allTags,created:'2026-10-26',source:'Only generated assets and real application APIs'};
fs.writeFileSync(path.join(run,'demo-state.json'),JSON.stringify(state,null,2));
console.log(JSON.stringify({libraries:libraries.length,people:people.length,resources:resources.length,tags:allTags.reduce((n,x)=>n+x.length,0)}));
