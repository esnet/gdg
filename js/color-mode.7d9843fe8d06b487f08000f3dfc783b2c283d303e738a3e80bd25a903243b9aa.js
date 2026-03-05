(()=>{(()=>{"use strict";let t=localStorage.getItem("theme"),d=()=>t||(window.matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light"),a=function(e){e==="auto"&&window.matchMedia("(prefers-color-scheme: dark)").matches?document.documentElement.setAttribute("data-bs-theme","dark"):document.documentElement.setAttribute("data-bs-theme",e),document.dispatchEvent(new CustomEvent("themeChanged",{detail:{theme:e}}))};a(d()),window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change",()=>{t!=="light"&&t!=="dark"&&a(d())}),window.addEventListener("DOMContentLoaded",()=>{document.querySelectorAll("[data-bs-theme-value]").forEach(e=>{e.addEventListener("click",()=>{let r=e.getAttribute("data-bs-theme-value");localStorage.setItem("theme",r),a(r)})})})})();})();
/*!
 * Modified from
 * Color mode toggler for Bootstrap's docs (https://getbootstrap.com/)
 * Copyright 2011-2022 The Bootstrap Authors
 * Licensed under the Creative Commons Attribution 3.0 Unported License.
 */
