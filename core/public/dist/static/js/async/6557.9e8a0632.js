"use strict";(self.webpackChunkfrontend=self.webpackChunkfrontend||[]).push([["6557"],{95352(e,t,r){r.d(t,{A:()=>o});var i=r(18123);let o=(0,i.defineComponent)({name:"Add",render:()=>(0,i.h)("svg",{width:"512",height:"512",viewBox:"0 0 512 512",fill:"none",xmlns:"http://www.w3.org/2000/svg"},(0,i.h)("path",{d:"M256 112V400M400 256H112",stroke:"currentColor","stroke-width":"32","stroke-linecap":"round","stroke-linejoin":"round"}))})},17766(e,t,r){r.d(t,{A:()=>C});var i=r(18123),o=r(12894),l=r(91945),n=r(47580),a=r(26005),s=r(88960),d=r(5423),u=r(64272),c=r(67872),p=r(635),h=r(61853),f=r(7647);let g={success:(0,i.h)(c.A,null),error:(0,i.h)(p.A,null),warning:(0,i.h)(h.A,null),info:(0,i.h)(f.A,null)},v=(0,i.defineComponent)({name:"ProgressCircle",props:{clsPrefix:{type:String,required:!0},status:{type:String,required:!0},strokeWidth:{type:Number,required:!0},fillColor:[String,Object],railColor:String,railStyle:[String,Object],percentage:{type:Number,default:0},offsetDegree:{type:Number,default:0},showIndicator:{type:Boolean,required:!0},indicatorTextColor:String,unit:String,viewBoxWidth:{type:Number,required:!0},gapDegree:{type:Number,required:!0},gapOffsetDegree:{type:Number,default:0}},setup(e,{slots:t}){let r=(0,i.computed)(()=>{let t="gradient",{fillColor:r}=e;return"object"==typeof r?`${t}-${(0,d.A)(JSON.stringify(r))}`:t});function o(t,i,o,l){let{gapDegree:n,viewBoxWidth:a,strokeWidth:s}=e,d=50+s/2,u=2*Math.PI*50;return{pathString:`M ${d},${d} m 0,50
      a 50,50 0 1 1 0,${-100}
      a 50,50 0 1 1 0,100`,pathStyle:{stroke:"rail"===l?o:"object"==typeof e.fillColor?`url(#${r.value})`:o,strokeDasharray:`${Math.min(t,100)/100*(u-n)}px ${8*a}px`,strokeDashoffset:`-${n/2}px`,transformOrigin:i?"center":void 0,transform:i?`rotate(${i}deg)`:void 0}}}return()=>{let l,n,a,{fillColor:s,railColor:d,strokeWidth:c,offsetDegree:p,status:h,percentage:f,showIndicator:v,indicatorTextColor:m,unit:y,gapOffsetDegree:b,clsPrefix:w}=e,{pathString:x,pathStyle:R}=o(100,0,d,"rail"),{pathString:B,pathStyle:C}=o(f,p,s,"fill"),k=100+c;return(0,i.h)("div",{class:`${w}-progress-content`,role:"none"},(0,i.h)("div",{class:`${w}-progress-graph`,"aria-hidden":!0},(0,i.h)("div",{class:`${w}-progress-graph-circle`,style:{transform:b?`rotate(${b}deg)`:void 0}},(0,i.h)("svg",{viewBox:`0 0 ${k} ${k}`},(n=(l="object"==typeof e.fillColor)?e.fillColor.stops[0]:"",a=l?e.fillColor.stops[1]:"",l&&(0,i.h)("defs",null,(0,i.h)("linearGradient",{id:r.value,x1:"0%",y1:"100%",x2:"100%",y2:"0%"},(0,i.h)("stop",{offset:"0%","stop-color":n}),(0,i.h)("stop",{offset:"100%","stop-color":a})))),(0,i.h)("g",null,(0,i.h)("path",{class:`${w}-progress-graph-circle-rail`,d:x,"stroke-width":c,"stroke-linecap":"round",fill:"none",style:R})),(0,i.h)("g",null,(0,i.h)("path",{class:[`${w}-progress-graph-circle-fill`,0===f&&`${w}-progress-graph-circle-fill--empty`],d:B,"stroke-width":c,"stroke-linecap":"round",fill:"none",style:C}))))),v?(0,i.h)("div",null,t.default?(0,i.h)("div",{class:`${w}-progress-custom-content`,role:"none"},t.default()):"default"!==h?(0,i.h)("div",{class:`${w}-progress-icon`,"aria-hidden":!0},(0,i.h)(u.A,{clsPrefix:w},{default:()=>g[h]})):(0,i.h)("div",{class:`${w}-progress-text`,style:{color:m},role:"none"},(0,i.h)("span",{class:`${w}-progress-text__percentage`},f),(0,i.h)("span",{class:`${w}-progress-text__unit`},y))):null)}}});var m=r(6280);let y={success:(0,i.h)(c.A,null),error:(0,i.h)(p.A,null),warning:(0,i.h)(h.A,null),info:(0,i.h)(f.A,null)},b=(0,i.defineComponent)({name:"ProgressLine",props:{clsPrefix:{type:String,required:!0},percentage:{type:Number,default:0},railColor:String,railStyle:[String,Object],fillColor:[String,Object],status:{type:String,required:!0},indicatorPlacement:{type:String,required:!0},indicatorTextColor:String,unit:{type:String,default:"%"},processing:{type:Boolean,required:!0},showIndicator:{type:Boolean,required:!0},height:[String,Number],railBorderRadius:[String,Number],fillBorderRadius:[String,Number]},setup(e,{slots:t}){let r=(0,i.computed)(()=>(0,m.i)(e.height)),o=(0,i.computed)(()=>{var t,r;return"object"==typeof e.fillColor?`linear-gradient(to right, ${null==(t=e.fillColor)?void 0:t.stops[0]} , ${null==(r=e.fillColor)?void 0:r.stops[1]})`:e.fillColor}),l=(0,i.computed)(()=>void 0!==e.railBorderRadius?(0,m.i)(e.railBorderRadius):void 0!==e.height?(0,m.i)(e.height,{c:.5}):""),n=(0,i.computed)(()=>void 0!==e.fillBorderRadius?(0,m.i)(e.fillBorderRadius):void 0!==e.railBorderRadius?(0,m.i)(e.railBorderRadius):void 0!==e.height?(0,m.i)(e.height,{c:.5}):"");return()=>{let{indicatorPlacement:a,railColor:s,railStyle:d,percentage:c,unit:p,indicatorTextColor:h,status:f,showIndicator:g,processing:v,clsPrefix:m}=e;return(0,i.h)("div",{class:`${m}-progress-content`,role:"none"},(0,i.h)("div",{class:`${m}-progress-graph`,"aria-hidden":!0},(0,i.h)("div",{class:[`${m}-progress-graph-line`,{[`${m}-progress-graph-line--indicator-${a}`]:!0}]},(0,i.h)("div",{class:`${m}-progress-graph-line-rail`,style:[{backgroundColor:s,height:r.value,borderRadius:l.value},d]},(0,i.h)("div",{class:[`${m}-progress-graph-line-fill`,v&&`${m}-progress-graph-line-fill--processing`],style:{maxWidth:`${e.percentage}%`,background:o.value,height:r.value,lineHeight:r.value,borderRadius:n.value}},"inside"===a?(0,i.h)("div",{class:`${m}-progress-graph-line-indicator`,style:{color:h}},t.default?t.default():`${c}${p}`):null)))),g&&"outside"===a?(0,i.h)("div",null,t.default?(0,i.h)("div",{class:`${m}-progress-custom-content`,style:{color:h},role:"none"},t.default()):"default"===f?(0,i.h)("div",{role:"none",class:`${m}-progress-icon ${m}-progress-icon--as-text`,style:{color:h}},c,p):(0,i.h)("div",{class:`${m}-progress-icon`,"aria-hidden":!0},(0,i.h)(u.A,{clsPrefix:m},{default:()=>y[f]}))):null)}}});function w(e,t,r=100){return`m ${r/2} ${r/2-e} a ${e} ${e} 0 1 1 0 ${2*e} a ${e} ${e} 0 1 1 0 -${2*e}`}let x=(0,i.defineComponent)({name:"ProgressMultipleCircle",props:{clsPrefix:{type:String,required:!0},viewBoxWidth:{type:Number,required:!0},percentage:{type:Array,default:[0]},strokeWidth:{type:Number,required:!0},circleGap:{type:Number,required:!0},showIndicator:{type:Boolean,required:!0},fillColor:{type:Array,default:()=>[]},railColor:{type:Array,default:()=>[]},railStyle:{type:Array,default:()=>[]}},setup(e,{slots:t}){let r=(0,i.computed)(()=>e.percentage.map((t,r)=>`${Math.PI*t/100*(e.viewBoxWidth/2-e.strokeWidth/2*(1+2*r)-e.circleGap*r)*2}, ${8*e.viewBoxWidth}`));return()=>{let{viewBoxWidth:o,strokeWidth:l,circleGap:n,showIndicator:a,fillColor:s,railColor:d,railStyle:u,percentage:c,clsPrefix:p}=e;return(0,i.h)("div",{class:`${p}-progress-content`,role:"none"},(0,i.h)("div",{class:`${p}-progress-graph`,"aria-hidden":!0},(0,i.h)("div",{class:`${p}-progress-graph-circle`},(0,i.h)("svg",{viewBox:`0 0 ${o} ${o}`},(0,i.h)("defs",null,c.map((t,r)=>{let o,l,n;return l="object"==typeof(o=e.fillColor[r])?o.stops[0]:"",n="object"==typeof o?o.stops[1]:"","object"==typeof e.fillColor[r]&&(0,i.h)("linearGradient",{id:`gradient-${r}`,x1:"100%",y1:"0%",x2:"0%",y2:"100%"},(0,i.h)("stop",{offset:"0%","stop-color":l}),(0,i.h)("stop",{offset:"100%","stop-color":n}))})),c.map((e,t)=>(0,i.h)("g",{key:t},(0,i.h)("path",{class:`${p}-progress-graph-circle-rail`,d:w(o/2-l/2*(1+2*t)-n*t,l,o),"stroke-width":l,"stroke-linecap":"round",fill:"none",style:[{strokeDashoffset:0,stroke:d[t]},u[t]]}),(0,i.h)("path",{class:[`${p}-progress-graph-circle-fill`,0===e&&`${p}-progress-graph-circle-fill--empty`],d:w(o/2-l/2*(1+2*t)-n*t,l,o),"stroke-width":l,"stroke-linecap":"round",fill:"none",style:{strokeDasharray:r.value[t],strokeDashoffset:0,stroke:"object"==typeof s[t]?`url(#gradient-${t})`:s[t]}})))))),a&&t.default?(0,i.h)("div",null,(0,i.h)("div",{class:`${p}-progress-text`},t.default())):null)}}}),R=(0,a.c)([(0,a.cB)("progress",{display:"inline-block"},[(0,a.cB)("progress-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 `),(0,a.cM)("line",`
 width: 100%;
 display: block;
 `,[(0,a.cB)("progress-content",`
 display: flex;
 align-items: center;
 `,[(0,a.cB)("progress-graph",{flex:1})]),(0,a.cB)("progress-custom-content",{marginLeft:"14px"}),(0,a.cB)("progress-icon",`
 width: 30px;
 padding-left: 14px;
 height: var(--n-icon-size-line);
 line-height: var(--n-icon-size-line);
 font-size: var(--n-icon-size-line);
 `,[(0,a.cM)("as-text",`
 color: var(--n-text-color-line-outer);
 text-align: center;
 width: 40px;
 font-size: var(--n-font-size);
 padding-left: 4px;
 transition: color .3s var(--n-bezier);
 `)])]),(0,a.cM)("circle, dashboard",{width:"120px"},[(0,a.cB)("progress-custom-content",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 justify-content: center;
 `),(0,a.cB)("progress-text",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 color: inherit;
 font-size: var(--n-font-size-circle);
 color: var(--n-text-color-circle);
 font-weight: var(--n-font-weight-circle);
 transition: color .3s var(--n-bezier);
 white-space: nowrap;
 `),(0,a.cB)("progress-icon",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 color: var(--n-icon-color);
 font-size: var(--n-icon-size-circle);
 `)]),(0,a.cM)("multiple-circle",`
 width: 200px;
 color: inherit;
 `,[(0,a.cB)("progress-text",`
 font-weight: var(--n-font-weight-circle);
 color: var(--n-text-color-circle);
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 justify-content: center;
 transition: color .3s var(--n-bezier);
 `)]),(0,a.cB)("progress-content",{position:"relative"}),(0,a.cB)("progress-graph",{position:"relative"},[(0,a.cB)("progress-graph-circle",[(0,a.c)("svg",{verticalAlign:"bottom"}),(0,a.cB)("progress-graph-circle-fill",`
 stroke: var(--n-fill-color);
 transition:
 opacity .3s var(--n-bezier),
 stroke .3s var(--n-bezier),
 stroke-dasharray .3s var(--n-bezier);
 `,[(0,a.cM)("empty",{opacity:0})]),(0,a.cB)("progress-graph-circle-rail",`
 transition: stroke .3s var(--n-bezier);
 overflow: hidden;
 stroke: var(--n-rail-color);
 `)]),(0,a.cB)("progress-graph-line",[(0,a.cM)("indicator-inside",[(0,a.cB)("progress-graph-line-rail",`
 height: 16px;
 line-height: 16px;
 border-radius: 10px;
 `,[(0,a.cB)("progress-graph-line-fill",`
 height: inherit;
 border-radius: 10px;
 `),(0,a.cB)("progress-graph-line-indicator",`
 background: #0000;
 white-space: nowrap;
 text-align: right;
 margin-left: 14px;
 margin-right: 14px;
 height: inherit;
 font-size: 12px;
 color: var(--n-text-color-line-inner);
 transition: color .3s var(--n-bezier);
 `)])]),(0,a.cM)("indicator-inside-label",`
 height: 16px;
 display: flex;
 align-items: center;
 `,[(0,a.cB)("progress-graph-line-rail",`
 flex: 1;
 transition: background-color .3s var(--n-bezier);
 `),(0,a.cB)("progress-graph-line-indicator",`
 background: var(--n-fill-color);
 font-size: 12px;
 transform: translateZ(0);
 display: flex;
 vertical-align: middle;
 height: 16px;
 line-height: 16px;
 padding: 0 10px;
 border-radius: 10px;
 position: absolute;
 white-space: nowrap;
 color: var(--n-text-color-line-inner);
 transition:
 right .2s var(--n-bezier),
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),(0,a.cB)("progress-graph-line-rail",`
 position: relative;
 overflow: hidden;
 height: var(--n-rail-height);
 border-radius: 5px;
 background-color: var(--n-rail-color);
 transition: background-color .3s var(--n-bezier);
 `,[(0,a.cB)("progress-graph-line-fill",`
 background: var(--n-fill-color);
 position: relative;
 border-radius: 5px;
 height: inherit;
 width: 100%;
 max-width: 0%;
 transition:
 background-color .3s var(--n-bezier),
 max-width .2s var(--n-bezier);
 `,[(0,a.cM)("processing",[(0,a.c)("&::after",`
 content: "";
 background-image: var(--n-line-bg-processing);
 animation: progress-processing-animation 2s var(--n-bezier) infinite;
 `)])])])])])]),(0,a.c)("@keyframes progress-processing-animation",`
 0% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 100%;
 opacity: 1;
 }
 66% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 0;
 opacity: 0;
 }
 100% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 0;
 opacity: 0;
 }
 `)]),B=Object.assign(Object.assign({},o.A.props),{processing:Boolean,type:{type:String,default:"line"},gapDegree:Number,gapOffsetDegree:Number,status:{type:String,default:"default"},railColor:[String,Array],railStyle:[String,Array],color:[String,Array,Object],viewBoxWidth:{type:Number,default:100},strokeWidth:{type:Number,default:7},percentage:[Number,Array],unit:{type:String,default:"%"},showIndicator:{type:Boolean,default:!0},indicatorPosition:{type:String,default:"outside"},indicatorPlacement:{type:String,default:"outside"},indicatorTextColor:String,circleGap:{type:Number,default:1},height:Number,borderRadius:[String,Number],fillBorderRadius:[String,Number],offsetDegree:Number}),C=(0,i.defineComponent)({name:"Progress",props:B,setup(e){let t=(0,i.computed)(()=>e.indicatorPlacement||e.indicatorPosition),r=(0,i.computed)(()=>e.gapDegree||0===e.gapDegree?e.gapDegree:"dashboard"===e.type?75:void 0),{mergedClsPrefixRef:d,inlineThemeDisabled:u}=(0,l.Ay)(e),c=(0,o.A)("Progress","-progress",R,s.A,e,d),p=(0,i.computed)(()=>{let{status:t}=e,{common:{cubicBezierEaseInOut:r},self:{fontSize:i,fontSizeCircle:o,railColor:l,railHeight:n,iconSizeCircle:s,iconSizeLine:d,textColorCircle:u,textColorLineInner:p,textColorLineOuter:h,lineBgProcessing:f,fontWeightCircle:g,[(0,a.cF)("iconColor",t)]:v,[(0,a.cF)("fillColor",t)]:m}}=c.value;return{"--n-bezier":r,"--n-fill-color":m,"--n-font-size":i,"--n-font-size-circle":o,"--n-font-weight-circle":g,"--n-icon-color":v,"--n-icon-size-circle":s,"--n-icon-size-line":d,"--n-line-bg-processing":f,"--n-rail-color":l,"--n-rail-height":n,"--n-text-color-circle":u,"--n-text-color-line-inner":p,"--n-text-color-line-outer":h}}),h=u?(0,n.R)("progress",(0,i.computed)(()=>e.status[0]),p,e):void 0;return{mergedClsPrefix:d,mergedIndicatorPlacement:t,gapDeg:r,cssVars:u?void 0:p,themeClass:null==h?void 0:h.themeClass,onRender:null==h?void 0:h.onRender}},render(){let{type:e,cssVars:t,indicatorTextColor:r,showIndicator:o,status:l,railColor:n,railStyle:a,color:s,percentage:d,viewBoxWidth:u,strokeWidth:c,mergedIndicatorPlacement:p,unit:h,borderRadius:f,fillBorderRadius:g,height:m,processing:y,circleGap:w,mergedClsPrefix:R,gapDeg:B,gapOffsetDegree:C,themeClass:k,$slots:$,onRender:P}=this;return null==P||P(),(0,i.h)("div",{class:[k,`${R}-progress`,`${R}-progress--${e}`,`${R}-progress--${l}`],style:t,"aria-valuemax":100,"aria-valuemin":0,"aria-valuenow":d,role:"circle"===e||"line"===e||"dashboard"===e?"progressbar":"none"},"circle"===e||"dashboard"===e?(0,i.h)(v,{clsPrefix:R,status:l,showIndicator:o,indicatorTextColor:r,railColor:n,fillColor:s,railStyle:a,offsetDegree:this.offsetDegree,percentage:d,viewBoxWidth:u,strokeWidth:c,gapDegree:void 0===B?75*("dashboard"===e):B,gapOffsetDegree:C,unit:h},$):"line"===e?(0,i.h)(b,{clsPrefix:R,status:l,showIndicator:o,indicatorTextColor:r,railColor:n,fillColor:s,railStyle:a,percentage:d,processing:y,indicatorPlacement:p,unit:h,fillBorderRadius:g,railBorderRadius:f,height:m},$):"multiple-circle"===e?(0,i.h)(x,{clsPrefix:R,strokeWidth:c,railColor:n,fillColor:s,railStyle:a,viewBoxWidth:u,percentage:d,showIndicator:o,circleGap:w},$):null)}})},11436(e,t,r){r.d(t,{A:()=>ei});var i=r(73321),o=r(11307),l=r(18123),n=r(12894),a=r(91945),s=r(12469),d=r(71811),u=r(47580),c=r(83032),p=r(75569),h=r(81270),f=r(54553),g=r(30560),v=r(7801),m=r(26005);let y=(0,m.c)([(0,m.cB)("upload","width: 100%;",[(0,m.cM)("dragger-inside",[(0,m.cB)("upload-trigger",`
 display: block;
 `)]),(0,m.cM)("drag-over",[(0,m.cB)("upload-dragger",`
 border: var(--n-dragger-border-hover);
 `)])]),(0,m.cB)("upload-dragger",`
 cursor: pointer;
 box-sizing: border-box;
 width: 100%;
 text-align: center;
 border-radius: var(--n-border-radius);
 padding: 24px;
 opacity: 1;
 transition:
 opacity .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 background-color: var(--n-dragger-color);
 border: var(--n-dragger-border);
 `,[(0,m.c)("&:hover",`
 border: var(--n-dragger-border-hover);
 `),(0,m.cM)("disabled",`
 cursor: not-allowed;
 `)]),(0,m.cB)("upload-trigger",`
 display: inline-block;
 box-sizing: border-box;
 opacity: 1;
 transition: opacity .3s var(--n-bezier);
 `,[(0,m.c)("+",[(0,m.cB)("upload-file-list","margin-top: 8px;")]),(0,m.cM)("disabled",`
 opacity: var(--n-item-disabled-opacity);
 cursor: not-allowed;
 `),(0,m.cM)("image-card",`
 width: 96px;
 height: 96px;
 `,[(0,m.cB)("base-icon",`
 font-size: 24px;
 `),(0,m.cB)("upload-dragger",`
 padding: 0;
 height: 100%;
 width: 100%;
 display: flex;
 align-items: center;
 justify-content: center;
 `)])]),(0,m.cB)("upload-file-list",`
 line-height: var(--n-line-height);
 opacity: 1;
 transition: opacity .3s var(--n-bezier);
 `,[(0,m.c)("a, img","outline: none;"),(0,m.cM)("disabled",`
 opacity: var(--n-item-disabled-opacity);
 cursor: not-allowed;
 `,[(0,m.cB)("upload-file","cursor: not-allowed;")]),(0,m.cM)("grid",`
 display: grid;
 grid-template-columns: repeat(auto-fill, 96px);
 grid-gap: 8px;
 margin-top: 0;
 `),(0,m.cB)("upload-file",`
 display: block;
 box-sizing: border-box;
 cursor: default;
 padding: 0px 12px 0 6px;
 transition: background-color .3s var(--n-bezier);
 border-radius: var(--n-border-radius);
 `,[(0,g._)(),(0,m.cB)("progress",[(0,g._)({foldPadding:!0})]),(0,m.c)("&:hover",`
 background-color: var(--n-item-color-hover);
 `,[(0,m.cB)("upload-file-info",[(0,m.cE)("action",`
 opacity: 1;
 `)])]),(0,m.cM)("image-type",`
 border-radius: var(--n-border-radius);
 text-decoration: underline;
 text-decoration-color: #0000;
 `,[(0,m.cB)("upload-file-info",`
 padding-top: 0px;
 padding-bottom: 0px;
 width: 100%;
 height: 100%;
 display: flex;
 justify-content: space-between;
 align-items: center;
 padding: 6px 0;
 `,[(0,m.cB)("progress",`
 padding: 2px 0;
 margin-bottom: 0;
 `),(0,m.cE)("name",`
 padding: 0 8px;
 `),(0,m.cE)("thumbnail",`
 width: 32px;
 height: 32px;
 font-size: 28px;
 display: flex;
 justify-content: center;
 align-items: center;
 `,[(0,m.c)("img",`
 width: 100%;
 `)])])]),(0,m.cM)("text-type",[(0,m.cB)("progress",`
 box-sizing: border-box;
 padding-bottom: 6px;
 margin-bottom: 6px;
 `)]),(0,m.cM)("image-card-type",`
 position: relative;
 width: 96px;
 height: 96px;
 border: var(--n-item-border-image-card);
 border-radius: var(--n-border-radius);
 padding: 0;
 display: flex;
 align-items: center;
 justify-content: center;
 transition: border-color .3s var(--n-bezier), background-color .3s var(--n-bezier);
 border-radius: var(--n-border-radius);
 overflow: hidden;
 `,[(0,m.cB)("progress",`
 position: absolute;
 left: 8px;
 bottom: 8px;
 right: 8px;
 width: unset;
 `),(0,m.cB)("upload-file-info",`
 padding: 0;
 width: 100%;
 height: 100%;
 `,[(0,m.cE)("thumbnail",`
 width: 100%;
 height: 100%;
 display: flex;
 flex-direction: column;
 align-items: center;
 justify-content: center;
 font-size: 36px;
 `,[(0,m.c)("img",`
 width: 100%;
 `)])]),(0,m.c)("&::before",`
 position: absolute;
 z-index: 1;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 border-radius: inherit;
 opacity: 0;
 transition: opacity .2s var(--n-bezier);
 content: "";
 `),(0,m.c)("&:hover",[(0,m.c)("&::before","opacity: 1;"),(0,m.cB)("upload-file-info",[(0,m.cE)("thumbnail","opacity: .12;")])])]),(0,m.cM)("error-status",[(0,m.c)("&:hover",`
 background-color: var(--n-item-color-hover-error);
 `),(0,m.cB)("upload-file-info",[(0,m.cE)("name","color: var(--n-item-text-color-error);"),(0,m.cE)("thumbnail","color: var(--n-item-text-color-error);")]),(0,m.cM)("image-card-type",`
 border: var(--n-item-border-image-card-error);
 `)]),(0,m.cM)("with-url",`
 cursor: pointer;
 `,[(0,m.cB)("upload-file-info",[(0,m.cE)("name",`
 color: var(--n-item-text-color-success);
 text-decoration-color: var(--n-item-text-color-success);
 `,[(0,m.c)("a",`
 text-decoration: underline;
 `)])])]),(0,m.cB)("upload-file-info",`
 position: relative;
 padding-top: 6px;
 padding-bottom: 6px;
 display: flex;
 flex-wrap: nowrap;
 `,[(0,m.cE)("thumbnail",`
 font-size: 18px;
 opacity: 1;
 transition: opacity .2s var(--n-bezier);
 color: var(--n-item-icon-color);
 `,[(0,m.cB)("base-icon",`
 margin-right: 2px;
 vertical-align: middle;
 transition: color .3s var(--n-bezier);
 `)]),(0,m.cE)("action",`
 padding-top: inherit;
 padding-bottom: inherit;
 position: absolute;
 right: 0;
 top: 0;
 bottom: 0;
 width: 80px;
 display: flex;
 align-items: center;
 transition: opacity .2s var(--n-bezier);
 justify-content: flex-end;
 opacity: 0;
 `,[(0,m.cB)("button",[(0,m.c)("&:not(:last-child)",{marginRight:"4px"}),(0,m.cB)("base-icon",[(0,m.c)("svg",[(0,v.N)()])])]),(0,m.cM)("image-type",`
 position: relative;
 max-width: 80px;
 width: auto;
 `),(0,m.cM)("image-card-type",`
 z-index: 2;
 position: absolute;
 width: 100%;
 height: 100%;
 left: 0;
 right: 0;
 bottom: 0;
 top: 0;
 display: flex;
 justify-content: center;
 align-items: center;
 `)]),(0,m.cE)("name",`
 color: var(--n-item-text-color);
 flex: 1;
 display: flex;
 justify-content: center;
 text-overflow: ellipsis;
 overflow: hidden;
 flex-direction: column;
 text-decoration-color: #0000;
 font-size: var(--n-font-size);
 transition:
 color .3s var(--n-bezier),
 text-decoration-color .3s var(--n-bezier); 
 `,[(0,m.c)("a",`
 color: inherit;
 text-decoration: underline;
 `)])])])]),(0,m.cB)("upload-file-input",`
 display: none;
 width: 0;
 height: 0;
 opacity: 0;
 `)]);var b=r(82777),w=r(38161),x=r(17150),R=r(93529),B=r(64272),C=r(53370),k=r(63834);let $=(0,k.l)("attach",()=>(0,l.h)("svg",{viewBox:"0 0 16 16",version:"1.1",xmlns:"http://www.w3.org/2000/svg"},(0,l.h)("g",{stroke:"none","stroke-width":"1",fill:"none","fill-rule":"evenodd"},(0,l.h)("g",{fill:"currentColor","fill-rule":"nonzero"},(0,l.h)("path",{d:"M3.25735931,8.70710678 L7.85355339,4.1109127 C8.82986412,3.13460197 10.4127766,3.13460197 11.3890873,4.1109127 C12.365398,5.08722343 12.365398,6.67013588 11.3890873,7.64644661 L6.08578644,12.9497475 C5.69526215,13.3402718 5.06209717,13.3402718 4.67157288,12.9497475 C4.28104858,12.5592232 4.28104858,11.9260582 4.67157288,11.5355339 L9.97487373,6.23223305 C10.1701359,6.0369709 10.1701359,5.72038841 9.97487373,5.52512627 C9.77961159,5.32986412 9.4630291,5.32986412 9.26776695,5.52512627 L3.96446609,10.8284271 C3.18341751,11.6094757 3.18341751,12.8758057 3.96446609,13.6568542 C4.74551468,14.4379028 6.01184464,14.4379028 6.79289322,13.6568542 L12.0961941,8.35355339 C13.4630291,6.98671837 13.4630291,4.77064094 12.0961941,3.40380592 C10.7293591,2.0369709 8.51328163,2.0369709 7.14644661,3.40380592 L2.55025253,8 C2.35499039,8.19526215 2.35499039,8.51184464 2.55025253,8.70710678 C2.74551468,8.90236893 3.06209717,8.90236893 3.25735931,8.70710678 Z"})))));var P=r(94254);let A=(0,k.l)("trash",()=>(0,l.h)("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 512 512"},(0,l.h)("path",{d:"M432,144,403.33,419.74A32,32,0,0,1,371.55,448H140.46a32,32,0,0,1-31.78-28.26L80,144",style:"fill: none; stroke: currentcolor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 32px;"}),(0,l.h)("rect",{x:"32",y:"64",width:"448",height:"80",rx:"16",ry:"16",style:"fill: none; stroke: currentcolor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 32px;"}),(0,l.h)("line",{x1:"312",y1:"240",x2:"200",y2:"352",style:"fill: none; stroke: currentcolor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 32px;"}),(0,l.h)("line",{x1:"312",y1:"352",x2:"200",y2:"240",style:"fill: none; stroke: currentcolor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 32px;"}))),z=(0,k.l)("cancel",()=>(0,l.h)("svg",{viewBox:"0 0 16 16",version:"1.1",xmlns:"http://www.w3.org/2000/svg"},(0,l.h)("g",{stroke:"none","stroke-width":"1",fill:"none","fill-rule":"evenodd"},(0,l.h)("g",{fill:"currentColor","fill-rule":"nonzero"},(0,l.h)("path",{d:"M2.58859116,2.7156945 L2.64644661,2.64644661 C2.82001296,2.47288026 3.08943736,2.45359511 3.2843055,2.58859116 L3.35355339,2.64644661 L8,7.293 L12.6464466,2.64644661 C12.8417088,2.45118446 13.1582912,2.45118446 13.3535534,2.64644661 C13.5488155,2.84170876 13.5488155,3.15829124 13.3535534,3.35355339 L8.707,8 L13.3535534,12.6464466 C13.5271197,12.820013 13.5464049,13.0894374 13.4114088,13.2843055 L13.3535534,13.3535534 C13.179987,13.5271197 12.9105626,13.5464049 12.7156945,13.4114088 L12.6464466,13.3535534 L8,8.707 L3.35355339,13.3535534 C3.15829124,13.5488155 2.84170876,13.5488155 2.64644661,13.3535534 C2.45118446,13.1582912 2.45118446,12.8417088 2.64644661,12.6464466 L7.293,8 L2.64644661,3.35355339 C2.47288026,3.17998704 2.45359511,2.91056264 2.58859116,2.7156945 L2.64644661,2.64644661 L2.58859116,2.7156945 Z"}))))),S=(0,k.l)("retry",()=>(0,l.h)("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 512 512"},(0,l.h)("path",{d:"M320,146s24.36-12-64-12A160,160,0,1,0,416,294",style:"fill: none; stroke: currentcolor; stroke-linecap: round; stroke-miterlimit: 10; stroke-width: 32px;"}),(0,l.h)("polyline",{points:"256 58 336 138 256 218",style:"fill: none; stroke: currentcolor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 32px;"})));var j=r(5375),O=r(29925),D=r(48209),T=r(92134);function L(){return(0,l.h)("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 28 28"},(0,l.h)("g",{fill:"none"},(0,l.h)("path",{d:"M21.75 3A3.25 3.25 0 0 1 25 6.25v15.5A3.25 3.25 0 0 1 21.75 25H6.25A3.25 3.25 0 0 1 3 21.75V6.25A3.25 3.25 0 0 1 6.25 3h15.5zm.583 20.4l-7.807-7.68a.75.75 0 0 0-.968-.07l-.084.07l-7.808 7.68c.183.065.38.1.584.1h15.5c.204 0 .4-.035.583-.1l-7.807-7.68l7.807 7.68zM21.75 4.5H6.25A1.75 1.75 0 0 0 4.5 6.25v15.5c0 .208.036.408.103.593l7.82-7.692a2.25 2.25 0 0 1 3.026-.117l.129.117l7.82 7.692c.066-.185.102-.385.102-.593V6.25a1.75 1.75 0 0 0-1.75-1.75zm-3.25 3a2.5 2.5 0 1 1 0 5a2.5 2.5 0 0 1 0-5zm0 1.5a1 1 0 1 0 0 2a1 1 0 0 0 0-2z",fill:"currentColor"})))}function M(){return(0,l.h)("svg",{xmlns:"http://www.w3.org/2000/svg",viewBox:"0 0 28 28"},(0,l.h)("g",{fill:"none"},(0,l.h)("path",{d:"M6.4 2A2.4 2.4 0 0 0 4 4.4v19.2A2.4 2.4 0 0 0 6.4 26h15.2a2.4 2.4 0 0 0 2.4-2.4V11.578c0-.729-.29-1.428-.805-1.944l-6.931-6.931A2.4 2.4 0 0 0 14.567 2H6.4zm-.9 2.4a.9.9 0 0 1 .9-.9H14V10a2 2 0 0 0 2 2h6.5v11.6a.9.9 0 0 1-.9.9H6.4a.9.9 0 0 1-.9-.9V4.4zm16.44 6.1H16a.5.5 0 0 1-.5-.5V4.06l6.44 6.44z",fill:"currentColor"})))}var F=r(17766);let U=(0,l.defineComponent)({name:"UploadProgress",props:{show:Boolean,percentage:{type:Number,required:!0},status:{type:String,required:!0}},setup:()=>({mergedTheme:(0,l.inject)(f.u).mergedThemeRef}),render(){return(0,l.h)(w.A,null,{default:()=>this.show?(0,l.h)(F.A,{type:"line",showIndicator:!1,percentage:this.percentage,status:this.status,height:2,theme:this.mergedTheme.peers.Progress,themeOverrides:this.mergedTheme.peerOverrides.Progress}):null})}});var E=r(12221),q=function(e,t,r,i){return new(r||(r=Promise))(function(o,l){function n(e){try{s(i.next(e))}catch(e){l(e)}}function a(e){try{s(i.throw(e))}catch(e){l(e)}}function s(e){var t;e.done?o(e.value):((t=e.value)instanceof r?t:new r(function(e){e(t)})).then(n,a)}s((i=i.apply(e,t||[])).next())})};function N(e){return e.includes("image/")}function _(e=""){let t=e.split("/"),r=t[t.length-1].split(/#|\?/)[0];return(/\.[^./\\]*$/.exec(r)||[""])[0]}let I=/(webp|svg|png|gif|jpg|jpeg|jfif|bmp|dpg|ico)$/i,H=e=>{if(e.type)return N(e.type);let t=_(e.name||"");if(I.test(t))return!0;let r=e.thumbnailUrl||e.url||"",i=_(r);return!!(/^data:image\//.test(r)||I.test(i))},W=E.B&&window.FileReader&&window.File;function X(e){let{id:t,name:r,percentage:i,status:o,url:l,file:n,thumbnailUrl:a,type:s,fullPath:d,batchId:u}=e;return{id:t,name:r,percentage:null!=i?i:null,status:o,url:null!=l?l:null,file:null!=n?n:null,thumbnailUrl:null!=a?a:null,type:null!=s?s:null,fullPath:null!=d?d:null,batchId:null!=u?u:null}}var G=function(e,t,r,i){return new(r||(r=Promise))(function(o,l){function n(e){try{s(i.next(e))}catch(e){l(e)}}function a(e){try{s(i.throw(e))}catch(e){l(e)}}function s(e){var t;e.done?o(e.value):((t=e.value)instanceof r?t:new r(function(e){e(t)})).then(n,a)}s((i=i.apply(e,t||[])).next())})};let V={paddingMedium:"0 3px",heightMedium:"24px",iconSizeMedium:"18px"},Y=(0,l.defineComponent)({name:"UploadFile",props:{clsPrefix:{type:String,required:!0},file:{type:Object,required:!0},listType:{type:String,required:!0},index:{type:Number,required:!0}},setup(e){let t=(0,l.inject)(f.u),r=(0,l.ref)(null),i=(0,l.ref)(""),o=(0,l.computed)(()=>{let{file:t}=e;return"finished"===t.status?"success":"error"===t.status?"error":"info"}),n=(0,l.computed)(()=>{let{file:t}=e;if("error"===t.status)return"error"}),a=(0,l.computed)(()=>{let{file:t}=e;return"uploading"===t.status}),s=(0,l.computed)(()=>{if(!t.showCancelButtonRef.value)return!1;let{file:r}=e;return["uploading","pending","error"].includes(r.status)}),d=(0,l.computed)(()=>{if(!t.showRemoveButtonRef.value)return!1;let{file:r}=e;return["finished"].includes(r.status)}),u=(0,l.computed)(()=>{if(!t.showDownloadButtonRef.value)return!1;let{file:r}=e;return["finished"].includes(r.status)}),p=(0,l.computed)(()=>{if(!t.showRetryButtonRef.value)return!1;let{file:r}=e;return["error"].includes(r.status)}),h=(0,R.A)(()=>i.value||e.file.thumbnailUrl||e.file.url),g=(0,l.computed)(()=>{if(!t.showPreviewButtonRef.value)return!1;let{file:{status:r},listType:i}=e;return["finished"].includes(r)&&h.value&&"image-card"===i});function v(r){let{xhrMap:i,doChange:o,onRemoveRef:{value:l},mergedFileListRef:{value:n}}=t;Promise.resolve(!l||l({file:Object.assign({},r),fileList:n,index:e.index})).then(e=>{if(!1===e)return;let t=Object.assign({},r,{status:"removed"});i.delete(r.id),o(t,void 0,{remove:!0})})}let m=()=>G(this,void 0,void 0,function*(){let{listType:r}=e;("image"===r||"image-card"===r)&&t.shouldUseThumbnailUrlRef.value(e.file)&&(i.value=yield t.getFileThumbnailUrlResolver(e.file))});return(0,l.watchEffect)(()=>{m()}),{mergedTheme:t.mergedThemeRef,progressStatus:o,buttonType:n,showProgress:a,disabled:t.mergedDisabledRef,showCancelButton:s,showRemoveButton:d,showDownloadButton:u,showRetryButton:p,showPreviewButton:g,mergedThumbnailUrl:h,shouldUseThumbnailUrl:t.shouldUseThumbnailUrlRef,renderIcon:t.renderIconRef,imageRef:r,handleRemoveOrCancelClick:function(r){r.preventDefault();let{file:i}=e;["finished","pending","error"].includes(i.status)?v(i):["uploading"].includes(i.status)?function(e){let{xhrMap:r}=t,i=r.get(e.id);null==i||i.abort(),v(Object.assign({},e))}(i):(0,c.R8)("upload","The button clicked type is unknown.")},handleDownloadClick:function(r){r.preventDefault(),function(e){let{onDownloadRef:{value:r},customDownloadRef:{value:i}}=t;Promise.resolve(!r||r(Object.assign({},e))).then(t=>{!1!==t&&(i?i(Object.assign({},e)):(0,O.R)(e.url,e.name))})}(e.file)},handleRetryClick:function(){return G(this,void 0,void 0,function*(){let r=t.onRetryRef.value;r&&!1===(yield r({file:e.file}))||t.submit(e.file.id)})},handlePreviewClick:function(i){let{onPreviewRef:{value:o}}=t;if(o)o(e.file,{event:i});else if("image-card"===e.listType){let{value:e}=r;if(!e)return;e.showPreview()}}}},render(){let e,{clsPrefix:t,mergedTheme:r,listType:i,file:o,renderIcon:n}=this,a="image"===i;e=a||"image-card"===i?this.shouldUseThumbnailUrl(o)&&this.mergedThumbnailUrl?(0,l.h)("a",{rel:"noopener noreferer",target:"_blank",href:o.url||void 0,class:`${t}-upload-file-info__thumbnail`,onClick:this.handlePreviewClick},"image-card"===i?(0,l.h)(T.A,{src:this.mergedThumbnailUrl||void 0,previewSrc:o.url||void 0,alt:o.name,ref:"imageRef"}):(0,l.h)("img",{src:this.mergedThumbnailUrl||void 0,alt:o.name})):(0,l.h)("span",{class:`${t}-upload-file-info__thumbnail`},n?n(o):H(o)?(0,l.h)(B.A,{clsPrefix:t},{default:L}):(0,l.h)(B.A,{clsPrefix:t},{default:M})):(0,l.h)("span",{class:`${t}-upload-file-info__thumbnail`},n?n(o):(0,l.h)(B.A,{clsPrefix:t},{default:()=>(0,l.h)($,null)}));let s=(0,l.h)(U,{show:this.showProgress,percentage:o.percentage||0,status:this.progressStatus});return(0,l.h)("div",{class:[`${t}-upload-file`,`${t}-upload-file--${this.progressStatus}-status`,o.url&&"error"!==o.status&&"image-card"!==i&&`${t}-upload-file--with-url`,`${t}-upload-file--${i}-type`]},(0,l.h)("div",{class:`${t}-upload-file-info`},e,(0,l.h)("div",{class:`${t}-upload-file-info__name`},("text"===i||"image"===i)&&(o.url&&"error"!==o.status?(0,l.h)("a",{rel:"noopener noreferer",target:"_blank",href:o.url||void 0,onClick:this.handlePreviewClick},o.name):(0,l.h)("span",{onClick:this.handlePreviewClick},o.name)),a&&s),(0,l.h)("div",{class:[`${t}-upload-file-info__action`,`${t}-upload-file-info__action--${i}-type`]},this.showPreviewButton?(0,l.h)(D.Ay,{key:"preview",quaternary:!0,type:this.buttonType,onClick:this.handlePreviewClick,theme:r.peers.Button,themeOverrides:r.peerOverrides.Button,builtinThemeOverrides:V},{icon:()=>(0,l.h)(B.A,{clsPrefix:t},{default:()=>(0,l.h)(P.A,null)})}):null,(this.showRemoveButton||this.showCancelButton)&&!this.disabled&&(0,l.h)(D.Ay,{key:"cancelOrTrash",theme:r.peers.Button,themeOverrides:r.peerOverrides.Button,quaternary:!0,builtinThemeOverrides:V,type:this.buttonType,onClick:this.handleRemoveOrCancelClick},{icon:()=>(0,l.h)(C.A,null,{default:()=>this.showRemoveButton?(0,l.h)(B.A,{clsPrefix:t,key:"trash"},{default:()=>(0,l.h)(A,null)}):(0,l.h)(B.A,{clsPrefix:t,key:"cancel"},{default:()=>(0,l.h)(z,null)})})}),this.showRetryButton&&!this.disabled&&(0,l.h)(D.Ay,{key:"retry",quaternary:!0,type:this.buttonType,onClick:this.handleRetryClick,theme:r.peers.Button,themeOverrides:r.peerOverrides.Button,builtinThemeOverrides:V},{icon:()=>(0,l.h)(B.A,{clsPrefix:t},{default:()=>(0,l.h)(S,null)})}),this.showDownloadButton?(0,l.h)(D.Ay,{key:"download",quaternary:!0,type:this.buttonType,onClick:this.handleDownloadClick,theme:r.peers.Button,themeOverrides:r.peerOverrides.Button,builtinThemeOverrides:V},{icon:()=>(0,l.h)(B.A,{clsPrefix:t},{default:()=>(0,l.h)(j.A,null)})}):null)),!a&&s)}});var Z=r(95352),J=r(80224);let K=(0,l.defineComponent)({name:"UploadTrigger",props:{abstract:Boolean},slots:Object,setup(e,{slots:t}){let r=(0,l.inject)(f.u,null);r||(0,c.$8)("upload-trigger","`n-upload-trigger` must be placed inside `n-upload`.");let{mergedClsPrefixRef:i,mergedDisabledRef:o,maxReachedRef:n,listTypeRef:a,dragOverRef:s,openOpenFileDialog:d,draggerInsideRef:u,handleFileAddition:p,mergedDirectoryDndRef:h,triggerClassRef:g,triggerStyleRef:v}=r,m=(0,l.computed)(()=>"image-card"===a.value);function y(){o.value||n.value||d()}function w(e){e.preventDefault(),s.value=!0}function x(e){e.preventDefault(),s.value=!0}function R(e){e.preventDefault(),s.value=!1}function C(e){var t;if(e.preventDefault(),!u.value||o.value||n.value){s.value=!1;return}let r=null==(t=e.dataTransfer)?void 0:t.items;(null==r?void 0:r.length)?(function(e,t){return q(this,void 0,void 0,function*(){let r=[];return yield function e(i){return q(this,void 0,void 0,function*(){for(let o of i)if(o){if(t&&o.isDirectory){let t,r=o.createReader(),i=[];try{do t=yield new Promise((e,t)=>{r.readEntries(e,t)}),i=i.concat(t);while(t.length>0)}catch(e){(0,c.z3)("upload","error happens when handling directory upload",e)}yield e(i)}else if(o.isFile)try{let e=yield new Promise((e,t)=>{o.file(e,t)});r.push({file:e,entry:o,source:"dnd"})}catch(e){(0,c.z3)("upload","error happens when handling file upload",e)}}})}(e),r})})(Array.from(r).map(e=>e.webkitGetAsEntry()),h.value).then(e=>{p(e)}).finally(()=>{s.value=!1}):s.value=!1}return()=>{var r;let{value:a}=i;return e.abstract?null==(r=t.default)?void 0:r.call(t,{handleClick:y,handleDrop:C,handleDragOver:w,handleDragEnter:x,handleDragLeave:R}):(0,l.h)("div",{class:[`${a}-upload-trigger`,(o.value||n.value)&&`${a}-upload-trigger--disabled`,m.value&&`${a}-upload-trigger--image-card`,g.value],style:v.value,onClick:y,onDrop:C,onDragover:w,onDragenter:x,onDragleave:R},m.value?(0,l.h)(b.A,null,{default:()=>(0,J.Nj)(t.default,()=>[(0,l.h)(B.A,{clsPrefix:a},{default:()=>(0,l.h)(Z.A,null)})])}):t)}}}),Q=(0,l.defineComponent)({name:"UploadFileList",setup(e,{slots:t}){let r=(0,l.inject)(f.u,null);r||(0,c.$8)("upload-file-list","`n-upload-file-list` must be placed inside `n-upload`.");let{abstractRef:i,mergedClsPrefixRef:o,listTypeRef:n,mergedFileListRef:a,fileListClassRef:s,fileListStyleRef:d,cssVarsRef:u,themeClassRef:p,maxReachedRef:h,showTriggerRef:g,imageGroupPropsRef:v}=r,m=(0,l.computed)(()=>"image-card"===n.value),y=()=>a.value.map((e,t)=>(0,l.h)(Y,{clsPrefix:o.value,key:e.id,file:e,index:t,listType:n.value}));return()=>{let{value:e}=o,{value:r}=i;return(0,l.h)("div",{class:[`${e}-upload-file-list`,m.value&&`${e}-upload-file-list--grid`,r?null==p?void 0:p.value:void 0,s.value],style:[r&&u?u.value:"",d.value]},m.value?(0,l.h)(x.Ay,Object.assign({},v.value),{default:y}):(0,l.h)(w.A,{group:!0},{default:y}),g.value&&!h.value&&m.value&&(0,l.h)(K,null,t))}}});var ee=function(e,t,r,i){return new(r||(r=Promise))(function(o,l){function n(e){try{s(i.next(e))}catch(e){l(e)}}function a(e){try{s(i.throw(e))}catch(e){l(e)}}function s(e){var t;e.done?o(e.value):((t=e.value)instanceof r?t:new r(function(e){e(t)})).then(n,a)}s((i=i.apply(e,t||[])).next())})};function et(e,t){return"function"==typeof e?e({file:t}):e||{}}let er=Object.assign(Object.assign({},n.A.props),{name:{type:String,default:"file"},accept:String,action:String,customRequest:Function,directory:Boolean,directoryDnd:{type:Boolean,default:void 0},method:{type:String,default:"POST"},multiple:Boolean,showFileList:{type:Boolean,default:!0},data:[Object,Function],headers:[Object,Function],withCredentials:Boolean,responseType:{type:String,default:""},disabled:{type:Boolean,default:void 0},onChange:Function,onRemove:Function,onFinish:Function,onError:Function,onRetry:Function,onBeforeUpload:Function,isErrorState:Function,onDownload:Function,customDownload:Function,defaultUpload:{type:Boolean,default:!0},fileList:Array,"onUpdate:fileList":[Function,Array],onUpdateFileList:[Function,Array],fileListClass:String,fileListStyle:[String,Object],defaultFileList:{type:Array,default:()=>[]},showCancelButton:{type:Boolean,default:!0},showRemoveButton:{type:Boolean,default:!0},showDownloadButton:Boolean,showRetryButton:{type:Boolean,default:!0},showPreviewButton:{type:Boolean,default:!0},listType:{type:String,default:"text"},onPreview:Function,shouldUseThumbnailUrl:{type:Function,default:e=>!!W&&H(e)},createThumbnailUrl:Function,abstract:Boolean,max:Number,showTrigger:{type:Boolean,default:!0},imageGroupProps:Object,inputProps:Object,triggerClass:String,triggerStyle:[String,Object],renderIcon:Function}),ei=(0,l.defineComponent)({name:"Upload",props:er,setup(e){e.abstract&&"image-card"===e.listType&&(0,c.$8)("upload","when the list-type is image-card, abstract is not supported.");let{mergedClsPrefixRef:t,inlineThemeDisabled:r,mergedRtlRef:g}=(0,a.Ay)(e),v=(0,n.A)("Upload","-upload",y,h.A,e,t),m=(0,s.I)("Upload",g,t),b=(0,d.A)(e),w=(0,l.ref)(e.defaultFileList),x=(0,l.toRef)(e,"fileList"),R=(0,l.ref)(null),B={value:!1},C=(0,l.ref)(!1),k=new Map,$=(0,o.A)(x,w),P=(0,l.computed)(()=>$.value.map(X)),A=(0,l.computed)(()=>{let{max:t}=e;return void 0!==t&&P.value.length>=t});function z(){var e;null==(e=R.value)||e.click()}let S=(0,l.computed)(()=>e.multiple||e.directory),j=(t,r,i={append:!1,remove:!1})=>{let{append:o,remove:l}=i,n=Array.from(P.value),a=n.findIndex(e=>e.id===t.id);if(o||l||~a){o?n.push(t):l?n.splice(a,1):n.splice(a,1,t);let{onChange:i}=e;i&&i({file:t,fileList:n,event:r}),function(t){let{"onUpdate:fileList":r,onUpdateFileList:i}=e;r&&(0,p.T)(r,t),i&&(0,p.T)(i,t),w.value=t}(n)}};function O(t,r){if(!t||0===t.length)return;let{onBeforeUpload:o}=e;t=S.value?t:[t[0]];let{max:n,accept:a}=e;t=t.filter(({file:e,source:t})=>{var r,i,o;return!("dnd"===t&&(null==a?void 0:a.trim()))||(r=e.name,i=e.type,o=a,r=r.toLowerCase(),i=i.toLocaleLowerCase(),(o=o.toLocaleLowerCase()).split(",").map(e=>e.trim()).filter(Boolean).some(e=>{if(e.startsWith(".")){if(r.endsWith(e))return!0}else{if(!e.includes("/"))return!0;let[t,r]=i.split("/"),[o,l]=e.split("/");if(("*"===o||t&&o&&o===t)&&("*"===l||r&&l&&l===r))return!0}return!1}))}),n&&(t=t.slice(0,n-P.value.length));let s=(0,i.sX)();Promise.all(t.map(e=>ee(this,[e],void 0,function*({file:e,entry:t}){var r;let l={id:(0,i.sX)(),batchId:s,name:e.name,status:"pending",percentage:0,file:e,url:null,type:e.type,thumbnailUrl:null,fullPath:null!=(r=null==t?void 0:t.fullPath)?r:`/${e.webkitRelativePath||e.name}`};return o&&(yield o({file:l,fileList:P.value}))===!1?null:l}))).then(e=>ee(this,void 0,void 0,function*(){let t=Promise.resolve();e.forEach(e=>{t=t.then(l.nextTick).then(()=>{e&&j(e,r,{append:!0})})}),yield t})).then(()=>{e.defaultUpload&&D()})}function D(t){let{method:r,action:i,withCredentials:o,headers:l,data:n,name:a}=e,s=void 0!==t?P.value.filter(e=>e.id===t):P.value,d=void 0!==t;s.forEach(t=>{let{status:s}=t;("pending"===s||"error"===s&&d)&&(e.customRequest?function(e){let{inst:t,file:r,data:i,headers:o,withCredentials:l,action:n,customRequest:a}=e,{doChange:s}=e.inst,d=0;a({file:r,data:i,headers:o,withCredentials:l,action:n,onProgress(e){let t=Object.assign({},r,{status:"uploading"}),i=e.percent;t.percentage=i,d=i,s(t)},onFinish(){var e;let i=Object.assign({},r,{status:"finished",percentage:d});s(i=X((null==(e=t.onFinish)?void 0:e.call(t,{file:i}))||i))},onError(){var e;let i=Object.assign({},r,{status:"error",percentage:d});s(i=X((null==(e=t.onError)?void 0:e.call(t,{file:i}))||i))}})}({inst:{doChange:j,xhrMap:k,onFinish:e.onFinish,onError:e.onError},file:t,action:i,withCredentials:o,headers:l,data:n,customRequest:e.customRequest}):function(e,t,r,{method:i,action:o,withCredentials:l,responseType:n,headers:a,data:s}){let d,u,c=new XMLHttpRequest;c.responseType=n,e.xhrMap.set(r.id,c),c.withCredentials=l;let p=new FormData;if((d=et(s,r))&&Object.keys(d).forEach(e=>{p.append(e,d[e])}),null!==r.file&&p.append(t,r.file),u=function(e,t,r){let{doChange:i,xhrMap:o}=e,l=0;function n(r){var n;let a=Object.assign({},t,{status:"error",percentage:l});o.delete(t.id),i(a=X((null==(n=e.onError)?void 0:n.call(e,{file:a,event:r}))||a),r)}return{handleXHRLoad:function(a){var s;if(e.isErrorState){if(e.isErrorState(r))return void n(a)}else if(r.status<200||r.status>=300)return void n(a);let d=Object.assign({},t,{status:"finished",percentage:l});o.delete(t.id),i(d=X((null==(s=e.onFinish)?void 0:s.call(e,{file:d,event:a}))||d),a)},handleXHRError:n,handleXHRAbort(e){let r=Object.assign({},t,{status:"removed",file:null,percentage:l});o.delete(t.id),i(r,e)},handleXHRProgress(e){let r=Object.assign({},t,{status:"uploading"});if(e.lengthComputable){let t=Math.ceil(e.loaded/e.total*100);r.percentage=t,l=t}i(r,e)}}}(e,r,c),c.onabort=u.handleXHRAbort,c.onerror=u.handleXHRError,c.onload=u.handleXHRLoad,c.upload&&(c.upload.onprogress=u.handleXHRProgress),void 0!==o){let t;c.open(i.toUpperCase(),o),(t=et(a,r))&&Object.keys(t).forEach(e=>{c.setRequestHeader(e,t[e])}),c.send(p);let l=Object.assign({},r,{status:"uploading"});e.doChange(l)}}({doChange:j,xhrMap:k,onFinish:e.onFinish,onError:e.onError,isErrorState:e.isErrorState},a,t,{method:r,action:i,withCredentials:o,responseType:e.responseType,headers:l,data:n}))})}let T=(0,l.computed)(()=>{let{common:{cubicBezierEaseInOut:e},self:{draggerColor:t,draggerBorder:r,draggerBorderHover:i,itemColorHover:o,itemColorHoverError:l,itemTextColorError:n,itemTextColorSuccess:a,itemTextColor:s,itemIconColor:d,itemDisabledOpacity:u,lineHeight:c,borderRadius:p,fontSize:h,itemBorderImageCardError:f,itemBorderImageCard:g}}=v.value;return{"--n-bezier":e,"--n-border-radius":p,"--n-dragger-border":r,"--n-dragger-border-hover":i,"--n-dragger-color":t,"--n-font-size":h,"--n-item-color-hover":o,"--n-item-color-hover-error":l,"--n-item-disabled-opacity":u,"--n-item-icon-color":d,"--n-item-text-color":s,"--n-item-text-color-error":n,"--n-item-text-color-success":a,"--n-line-height":c,"--n-item-border-image-card-error":f,"--n-item-border-image-card":g}}),L=r?(0,u.R)("upload",void 0,T,e):void 0;return(0,l.provide)(f.u,{mergedClsPrefixRef:t,mergedThemeRef:v,showCancelButtonRef:(0,l.toRef)(e,"showCancelButton"),showDownloadButtonRef:(0,l.toRef)(e,"showDownloadButton"),showRemoveButtonRef:(0,l.toRef)(e,"showRemoveButton"),showRetryButtonRef:(0,l.toRef)(e,"showRetryButton"),onRemoveRef:(0,l.toRef)(e,"onRemove"),onDownloadRef:(0,l.toRef)(e,"onDownload"),customDownloadRef:(0,l.toRef)(e,"customDownload"),mergedFileListRef:P,triggerClassRef:(0,l.toRef)(e,"triggerClass"),triggerStyleRef:(0,l.toRef)(e,"triggerStyle"),shouldUseThumbnailUrlRef:(0,l.toRef)(e,"shouldUseThumbnailUrl"),renderIconRef:(0,l.toRef)(e,"renderIcon"),xhrMap:k,submit:D,doChange:j,showPreviewButtonRef:(0,l.toRef)(e,"showPreviewButton"),onPreviewRef:(0,l.toRef)(e,"onPreview"),getFileThumbnailUrlResolver:function(t){var r;if(t.thumbnailUrl)return t.thumbnailUrl;let{createThumbnailUrl:i}=e;return i?null!=(r=i(t.file,t))?r:t.url||"":t.url?t.url:t.file?function(e){return q(this,void 0,void 0,function*(){return yield new Promise(t=>{t(e.type&&N(e.type)?window.URL.createObjectURL(e):"")})})}(t.file):""},listTypeRef:(0,l.toRef)(e,"listType"),dragOverRef:C,openOpenFileDialog:z,draggerInsideRef:B,handleFileAddition:O,mergedDisabledRef:b.mergedDisabledRef,maxReachedRef:A,fileListClassRef:(0,l.toRef)(e,"fileListClass"),fileListStyleRef:(0,l.toRef)(e,"fileListStyle"),abstractRef:(0,l.toRef)(e,"abstract"),acceptRef:(0,l.toRef)(e,"accept"),cssVarsRef:r?void 0:T,themeClassRef:null==L?void 0:L.themeClass,onRender:null==L?void 0:L.onRender,showTriggerRef:(0,l.toRef)(e,"showTrigger"),imageGroupPropsRef:(0,l.toRef)(e,"imageGroupProps"),mergedDirectoryDndRef:(0,l.computed)(()=>{var t;return null!=(t=e.directoryDnd)?t:e.directory}),onRetryRef:(0,l.toRef)(e,"onRetry")}),Object.assign({mergedClsPrefix:t,draggerInsideRef:B,rtlEnabled:m,inputElRef:R,mergedTheme:v,dragOver:C,mergedMultiple:S,cssVars:r?void 0:T,themeClass:null==L?void 0:L.themeClass,onRender:null==L?void 0:L.onRender,handleFileInputChange:function(e){let t=e.target;O(t.files?Array.from(t.files).map(e=>({file:e,entry:null,source:"input"})):null,e),t.value=""}},{clear:()=>{w.value=[]},submit:D,openOpenFileDialog:z})},render(){var e,t;let{draggerInsideRef:r,mergedClsPrefix:i,$slots:o,directory:n,onRender:a}=this;if(o.default&&!this.abstract){let t=o.default()[0];(null==(e=null==t?void 0:t.type)?void 0:e[b.P])&&(r.value=!0)}let s=(0,l.h)("input",Object.assign({},this.inputProps,{ref:"inputElRef",type:"file",class:`${i}-upload-file-input`,accept:this.accept,multiple:this.mergedMultiple,onChange:this.handleFileInputChange,webkitdirectory:n||void 0,directory:n||void 0}));return this.abstract?(0,l.h)(l.Fragment,null,null==(t=o.default)?void 0:t.call(o),(0,l.h)(l.Teleport,{to:"body"},s)):(null==a||a(),(0,l.h)("div",{class:[`${i}-upload`,this.rtlEnabled&&`${i}-upload--rtl`,r.value&&`${i}-upload--dragger-inside`,this.dragOver&&`${i}-upload--drag-over`,this.themeClass],style:this.cssVars},s,this.showTrigger&&"image-card"!==this.listType&&(0,l.h)(K,null,o),this.showFileList&&(0,l.h)(Q,null,o)))}})},82777(e,t,r){r.d(t,{A:()=>a,P:()=>n});var i=r(18123),o=r(83032),l=r(54553);let n="__UPLOAD_DRAGGER__",a=(0,i.defineComponent)({name:"UploadDragger",[n]:!0,setup(e,{slots:t}){let r=(0,i.inject)(l.u,null);return r||(0,o.$8)("upload-dragger","`n-upload-dragger` must be placed inside `n-upload`."),()=>{let{mergedClsPrefixRef:{value:e},mergedDisabledRef:{value:o},maxReachedRef:{value:l}}=r;return(0,i.h)("div",{class:[`${e}-upload-dragger`,(o||l)&&`${e}-upload-dragger--disabled`]},t)}}})},54553(e,t,r){r.d(t,{u:()=>i});let i=(0,r(98667).D)("n-upload")}}]);