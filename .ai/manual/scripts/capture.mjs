import fs from 'node:fs';
import path from 'node:path';
import {api,browser,home,pageReady,shot,base,state} from './browser.mjs';
const phase=process.argv[2]||'layouts';
const b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});
const errors=[];p.on('pageerror',e=>errors.push(e.message));
const lib=state.libraries[0].id;
const original=JSON.parse(await api('/filesBases/config/'+lib+'/filesBases')||'{}');
async function setDisplay(values){await api('/filesBases/setConfig/filesBases',{id:lib,config:JSON.stringify({...original,...values})},'PUT');}
try {
 if(phase==='layouts'){
  const modes=['coverPoster','coverPosterCinemaGallery','coverPosterBox','coverPosterBoxWideSeparate','coverPosterSimple','coverPosterSimpleExpand','coverPosterWaterfall','coverPosterMosaic','coverPosterMosaicShortVideo','coverPosterCompactWall','shortVideo','shortVideoTopBottom','table'];
  for(const mode of modes){await setDisplay({resourcesShowMode:mode,coverPosterWaterfallColumn:6});await home(p);await shot(p,'layout-'+mode);}
  await setDisplay({resourcesShowMode:'coverPoster'});await home(p);
  await p.getByTitle('主页样式：经典布局').click();await p.getByText('画室布局',{exact:true}).click();await pageReady(p);await shot(p,'home-studio');
  const config=await api('/app/getConfig');config.homeMode='classic';await api('/app/setConfig',config,'PUT');
 } else if(phase==='tags'){
  await setDisplay({resourcesShowMode:'coverPoster'});await home(p);
  await p.locator('.content-right-click-menu').first().click({button:'right'});await p.getByText('打标签',{exact:true}).click();
  await p.getByRole('dialog',{name:'设置资源标签'}).waitFor();await shot(p,'tag-single');
  const singleChoice=p.getByRole('dialog',{name:'设置资源标签'}).locator('label.el-checkbox').filter({hasText:'精选'});
  if(!await singleChoice.locator('input').isChecked())await singleChoice.click();
  await p.getByRole('dialog',{name:'设置资源标签'}).getByRole('button',{name:'提交',exact:true}).click();await pageReady(p);
  const changed=await api('/resource/info/'+state.resources[0].id);if(!changed.tags.some(x=>x.name==='精选'))throw Error('single tagging did not persist');console.log('verified: single tag persists');
  await p.getByTitle('开启管理模式').click();await pageReady(p);
  for(let i=0;i<3;i++)await p.locator('.el-table__body-wrapper input[type=checkbox]').nth(i).evaluate(el=>el.click());
  console.log('checked rows',await p.locator('.el-table__body-wrapper input:checked').count());await shot(p,'management');
  // Background list refreshes can replace table rows while the screenshot is
  // settling. Reconfirm the real selection immediately before the action.
  for(let i=0;i<3;i++){const box=p.locator('.el-table__body-wrapper input[type=checkbox]').nth(i);if(!await box.isChecked())await box.evaluate(el=>el.click());}
  await p.getByRole('button',{name:'批量添加标签',exact:true}).click();await p.getByRole('dialog',{name:'批量添加标签'}).locator('label.el-checkbox').filter({hasText:'学习参考'}).click();await shot(p,'tag-batch');
  await p.getByRole('dialog',{name:'批量添加标签'}).getByRole('button',{name:'提交',exact:true}).click();await pageReady(p);
  console.log('verified: batch tag submitted');
  await home(p);await p.getByTitle('标签管理',{exact:true}).click();await shot(p,'tag-manage');
 } else if(phase==='settings'){
  await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await pageReady(p);
  for(const [label,name]of [['软件设置','general'],['文件数据库','libraries'],['演员数据库','performer-libraries'],['演员头像库','avatars'],['标签数据库','tag-database'],['AI 自动标签','ai-tags'],['视频信息采集','metadata'],['视频转码与剪辑','transcode'],['计划任务','cron'],['视频重复检测','duplicates'],['TVBox推荐','tvbox'],['数据库清理','cleanup'],['自动备份','backup']]){
   await p.getByRole('tab',{name:label,exact:true}).click();if(['libraries','tag-database'].includes(name))await p.getByRole('tab',{name:'山海影像',exact:true}).click();await shot(p,'setting-'+name);
  }
 } else if(phase==='resources'){
  await home(p);await p.locator('.content-right-click-menu').first().click({button:'right'});await shot(p,'resource-menu');await p.getByText('编辑',{exact:true}).click();await shot(p,'resource-edit');
  await home(p);await p.getByTitle('导入与刮削').click();await shot(p,'import');
  await p.getByText('刮削',{exact:true}).first().click();await shot(p,'scraper');
  await home(p);await p.getByTitle('演员',{exact:true}).click();await shot(p,'performers');
  await p.goto('http://127.0.0.1:18346/moments',{waitUntil:'networkidle'});await shot(p,'moments');
  await p.goto('http://127.0.0.1:18346/mobile',{waitUntil:'networkidle'});await shot(p,'mobile-wide');
  await p.setViewportSize({width:430,height:932});await p.reload({waitUntil:'networkidle'});await shot(p,'mobile');
 }
 console.log('pageErrors',JSON.stringify(errors));
}catch(e){console.error(e);console.log((await p.locator('body').innerText()).slice(-5000));await shot(p,'debug-'+phase);process.exitCode=1;}
finally{await b.close();}
// These layouts need landscape photography and their own width/height setup.
// Always recapture them last so the general portrait pass cannot overwrite them.
if(phase==='layouts'&&!process.exitCode)await import('./capture_wide_layouts.mjs');
