self.addEventListener('install', (e) => {
    console.log('[Service Worker] Install');
});

self.addEventListener('fetch', (e) => {
    // Nanti kita isi logic offline disini, sementara pass through aja
});