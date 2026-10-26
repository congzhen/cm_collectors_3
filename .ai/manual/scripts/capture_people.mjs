import {api,browser,home,pageReady,shot,state} from './browser.mjs';
const b=await browser();const p=await b.newPage({viewport:{width:1440,height:1000},deviceScaleFactor:1});
try{
 await p.goto('http://127.0.0.1:18346/performer/'+state.perBase,{waitUntil:'networkidle'});await pageReady(p);
 await p.getByRole('slider').last().focus();await p.keyboard.press('End');await shot(p,'performers');
 await p.goto('http://127.0.0.1:18346/setting',{waitUntil:'networkidle'});await p.getByRole('tab',{name:'演员数据库',exact:true}).click();await pageReady(p);await p.getByRole('slider').last().focus();await p.keyboard.press('End');await shot(p,'setting-performer-libraries');
 await p.goto('http://127.0.0.1:18346/performer/tags/'+state.perBase,{waitUntil:'networkidle'});await shot(p,'performer-tags');
 const lib=state.libraries[0].id;const config=JSON.parse(await api('/filesBases/config/'+lib+'/filesBases'));
 await api('/filesBases/setConfig/filesBases',{id:lib,config:JSON.stringify({...config,resourcesShowMode:'coverPosterWaterfall'})},'PUT');
 await home(p);await p.getByRole('slider').last().focus();await p.keyboard.press('ArrowLeft');await p.keyboard.press('ArrowLeft');await shot(p,'layout-coverPosterWaterfall');
 await api('/filesBases/setConfig/filesBases',{id:lib,config:JSON.stringify({...config,resourcesShowMode:'coverPoster'})},'PUT');
}finally{await b.close();}
