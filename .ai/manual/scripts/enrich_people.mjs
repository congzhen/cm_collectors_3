import fs from 'node:fs';
import path from 'node:path';
import {api,state,run} from './browser.mjs';
const marker=path.join(run,'people-enriched.json');
if(fs.existsSync(marker)){console.log('People tags reused');process.exit(0);}
const groups={'创作领域':['自然影像','旅行记录','城市观察','人文访谈','建筑摄影','生活记录','海岸摄影','山野行走','森林观察','湖泊记录','夜景摄影','街巷故事','纪录创作','影像剪辑','声音设计','视觉设计','插画创作','绘本创作','景观记录','天文观察','美食记录','文化研究','地理探索','手工制作'],'所属团队':['山海工作室','远行小组','城市观察组','光影工坊','森林小队','海岸计划','星河团队','生活研究所'],'个人偏好':['优先关注','合作创作','新发现','作品待看','长期关注','灵感来源','学习参考','精选作者']};
const pools=[];
for(const [name,names]of Object.entries(groups)){
 const cls=await api('/performerTagClass/create',{performerBases_id:state.perBase,name,sort:pools.length,status:true});
 const tags=[];for(let i=0;i<names.length;i++)tags.push(await api('/performerTag/create',{performerTagClass_id:cls.id,name:names[i],sort:i,status:true}));pools.push(tags);
}
for(let i=0;i<state.people.length;i++)await api('/performer/batchSetTags',{performerIds:[state.people[i].id],performerBasesId:state.perBase,tagIds:pools.map(pool=>pool[i%pool.length].id),mode:'replace'},'PUT');
fs.writeFileSync(marker,JSON.stringify({classes:3,tags:40,people:24}));console.log('3 people categories, 40 tags, 24 people assigned');
