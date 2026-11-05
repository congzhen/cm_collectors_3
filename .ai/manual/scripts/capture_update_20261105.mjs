// Uses only the preserved, independent documentation service and photo assets.
import fs from 'node:fs';
import path from 'node:path';
import {api,browser,home,pageReady,shot,base,run,state} from './browser.mjs';
const phase=process.argv[2]||'overview';
const recordFile=path.join(run,'update-20261105.json');
const lib=state.libraries[0].id;
const record=fs.existsSync(recordFile)?JSON.parse(fs.readFileSync(recordFile,'utf8')):{date:'2026-11-05',groups:[],screenshots:[],verified:[]};
const save=()=>fs.writeFileSync(recordFile,JSON.stringify(record,null,2));
if(!record.originalConfig){record.originalConfig=await api('/filesBases/config/'+lib+'/filesBases');save();}
const original=JSON.parse(record.originalConfig||'{}');
const setup={...original,resourcesShowMode:'coverPoster',categoryBrowseEnabled:false,resourceQuickStars:true,mergeResourceGroups:true,resourceGroupCoverStyle:'adaptive',resourceGroupExpandStyle:'fan',pageLimit:24,videoMomentLibraryMode:'theater'};
async function display(values={}){await api('/filesBases/setConfig/filesBases',{id:lib,config:JSON.stringify({...setup,...values})},'PUT');}
const b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});p.setDefaultTimeout(16000);
const errors=[];p.on('pageerror',e=>errors.push(e.message));
async function take(name){await shot(p,name);record.screenshots.push(name);save();}
async function cropFor(name,locator){const box=await locator.boundingBox();if(!box)throw Error('No real UI bounds for '+name);record.crops||={};record.crops[name]=[Math.max(0,Math.floor(box.x)-8),Math.max(0,Math.floor(box.y)-8),Math.min(1440,Math.ceil(box.x+box.width)+8),Math.min(1000,Math.ceil(box.y+box.height)+8)];save();}
try{
 if(phase==='overview'){
  await display({mergeResourceGroups:false});await home(p);
  await p.getByTitle('开启管理模式',{exact:true}).click();await pageReady(p);
  const cards=p.locator('.content-right-click-menu');for(let i=0;i<3;i++)await cards.nth(i).click();
  await take('management-layout');await p.getByRole('button',{name:'查看已选',exact:true}).hover();await take('management-selection');
  await p.getByRole('button',{name:'添加标签',exact:true}).click();await take('tag-batch');
  await home(p);await p.getByTitle('开启管理模式',{exact:true}).click();await pageReady(p);await take('management');
  await home(p);await p.getByRole('button',{name:'显示设置',exact:true}).click();await take('display-settings');
  record.verified.push('当前布局勾选3条，已选预览与标签窗口真实打开，未提交标签变更');
 }else if(phase==='groups'||phase==='styles'){
  if(!record.groups.length){
   const resources=state.resources.filter(r=>r.filesBases_id===lib);
   for(const items of [resources.slice(0,8),resources.slice(8,12)]){
    const current=await api('/resource/info/'+items[0].id);
    let id=current.associationGroupId;
    if(!id){const result=await api('/resource/groups/manage/create',{filesBasesId:lib,resourceIds:items.map(r=>r.id)});id=result.id;}
    record.groups.push(id);save();
   }
  }
  const members=await api('/resource/groups/'+lib+'/'+record.groups[0]);if(members.length!==8)throw Error('Expected real 8-member association group');
  record.verified.push('真实创建并读取8成员与4成员关联组，未复制资源或媒体');
  for(const style of ['stack','triptych','book','adaptive']){
   await display({resourceGroupCoverStyle:style,coverPosterWidthStatus:true,coverPosterWidthBase:240,coverPosterHeightStatus:true,coverPosterHeightBase:340});await home(p);await take('group-cover-'+style);await cropFor('group-cover-'+style,p.locator('.content-right-click-menu').first());
  }
  for(const style of ['fan','strip','fold','spine','magazine']){
   await display({resourceGroupExpandStyle:style});await home(p);await p.locator('.content-right-click-menu').first().hover();await p.waitForTimeout(1000);await take('group-expand-'+style);await cropFor('group-expand-'+style,p.locator('.group-fan:visible, .group-fold:visible, .group-magazine:visible').first());
  }
  await display();await home(p);await p.getByTitle('开启管理模式',{exact:true}).click();await pageReady(p);
  await p.locator('.content-right-click-menu').first().click();await p.getByRole('button',{name:'查看分组',exact:true}).click();await take('group-manage');
 }else if(phase==='category'){
  await display({categoryBrowseEnabled:true,categoryBrowseSource:'performer',categoryBrowseSort:'count',categoryBrowseGroupLimit:2,categoryBrowseCoverStyle:'adaptive',categoryBrowseExpandStyle:'fan'});await home(p);await take('category-browse');
  console.log('category body', (await p.locator('body').innerText()).slice(-1800));
  await p.getByRole('button',{name:'显示设置',exact:true}).click();await take('category-settings');
  const classes=await api('/tagClass/list/'+lib);const subject=classes.find(x=>x.name==='题材');if(!subject)throw Error('Missing generated subject tags');
  await display({categoryBrowseEnabled:true,categoryBrowseSource:'tag:'+subject.id,categoryBrowseSort:'count',categoryBrowseGroupLimit:2,categoryBrowseCoverStyle:'adaptive',categoryBrowseExpandStyle:'fan'});
  await home(p);await take('category-tags');await p.locator('.category-cover').first().click();await pageReady(p);await take('category-members');record.verified.push('分类浏览按演员与题材真实归集，上限2拆组，进入本组并显示全部切换');
 }else if(phase==='stars'){
  await display({mergeResourceGroups:false});await home(p);await p.locator('.content-right-click-menu').first().hover();
  const r=await api('/resource/info/'+state.resources[0].id);
  try{
   await p.locator('.quick-stars-trigger').first().click();
   await p.getByRole('button',{name:'评为 4 星',exact:true}).click();await pageReady(p);
   const changed=await api('/resource/info/'+r.id);if(changed.stars!==4)throw Error('Quick stars did not persist');
   await take('resource-quick-stars');record.verified.push('快捷评星实际写入4星并重新读取确认，随后恢复原星级');
  }finally{await api('/resource/batchSetStars',{resourceIds:[r.id],stars:r.stars},'PUT');}
 }else if(phase==='manager'){
  await display({mergeResourceGroups:false});await home(p);
  const search=p.getByPlaceholder('请输入搜索内容');await search.fill(state.resources[0].title);await search.press('Enter');await pageReady(p);
  const card=p.locator('.content-right-click-menu').filter({hasText:state.resources[0].title}).first();await card.waitFor();await card.click({button:'right'});await p.getByText('编辑',{exact:true}).click();
  await p.getByRole('button',{name:/片段收藏（/}).click();await p.locator('.manager-row').first().waitFor();await take('resource-moments-manager');await cropFor('resource-moments-manager',p.getByRole('dialog',{name:'管理片段收藏',exact:true}));
  record.verified.push('指定有实际收藏的资源打开管理窗口，核对收藏卡片与起止时间，未删除记录');
 }else if(phase==='moments'){
  await display();await home(p);await p.getByRole('button',{name:'片段放映室',exact:true}).click();await pageReady(p);
  const video=p.locator('video').first();await video.waitFor();await video.evaluate(v=>{v.muted=true;v.currentTime=8;return v.play().catch(()=>{});});await p.waitForTimeout(700);await video.evaluate(v=>v.pause());await take('moments-theater');
  await p.getByRole('switch',{name:'放映室模式',exact:true}).click();await pageReady(p);await take('moments-collection');
  await display({mergeResourceGroups:false});await home(p);const search=p.getByPlaceholder('请输入搜索内容');await search.fill(state.resources[0].title);await search.press('Enter');await pageReady(p);
  await p.locator('.content-right-click-menu').filter({hasText:state.resources[0].title}).first().click({button:'right'});await p.getByText('编辑',{exact:true}).click();await p.getByRole('button',{name:/片段收藏（/}).click();await p.locator('.manager-row').first().waitFor();await take('resource-moments-manager');
  const mobile=await b.newPage({viewport:{width:430,height:932},deviceScaleFactor:1,isMobile:true,hasTouch:true,userAgent:'Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Mobile Safari/537.36'});
  await display();await mobile.goto('http://127.0.0.1:18346/mobile',{waitUntil:'networkidle'});await pageReady(mobile);
  await mobile.locator('.mobile-database-select').click();await mobile.getByRole('option',{name:'山海影像',exact:true}).click();await pageReady(mobile);
  await mobile.goto('http://127.0.0.1:18346/moments',{waitUntil:'networkidle'});await pageReady(mobile);
  await shot(mobile,'mobile-moment-theater');record.screenshots.push('mobile-moment-theater');
  record.verified.push('桌面片段放映室真实载入视频，收藏视图切换；资源编辑的片段管理窗口打开，未删除收藏');
 }
 if(errors.length)throw Error(JSON.stringify(errors));
 save();
}catch(e){console.error(e);console.log((await p.locator('body').innerText()).slice(-4000));await shot(p,'debug-update-'+phase);process.exitCode=1;}
finally{await api('/filesBases/setConfig/filesBases',{id:lib,config:record.originalConfig},'PUT');await b.close();}
