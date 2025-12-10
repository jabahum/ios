// Bootstrap CSS fallback - improved offline detection and handling
(function() {
    'use strict';
    
    // Check if online (for CDN fallback)
    function isOnline() {
        try {
            return navigator.onLine !== false && 
                   (typeof navigator.onLine === 'undefined' || navigator.onLine);
        } catch(e) {
            // Assume offline if we can't determine
            return false;
        }
    }
    
    // Try alternative local paths for Bootstrap
    function tryLocalBootstrapPaths(callback) {
        var localPaths = [
            '/static/auth/libs/bootstrap/css/bootstrap.min.css',
            '/static/app/landing/css/bootstrap_4.1.3.min.css',
            '/static/auth/css/bootstrap.min.css'
        ];
        
        var currentPathIndex = 0;
        var tried = false;
        
        function tryNextPath() {
            if (currentPathIndex >= localPaths.length) {
                if (callback) callback(false); // All local paths failed
                return;
            }
            
            var path = localPaths[currentPathIndex];
            currentPathIndex++;
            
            // Skip if this is the primary path (already tried)
            if (path === '/static/css/bootstrap.min.css') {
                tryNextPath();
                return;
            }
            
            // Check if this path is already loaded
            var existingLinks = document.querySelectorAll('link[href="' + path + '"]');
            if (existingLinks.length > 0) {
                // Already loaded, success!
                console.info('Bootstrap: Already loaded from:', path);
                if (callback) callback(true);
                return;
            }
            
            var link = document.createElement('link');
            link.rel = 'stylesheet';
            link.href = path;
            link.id = 'bootstrap-css-local-fallback';
            
            link.onload = function() {
                console.info('Bootstrap: Successfully loaded from local path:', path);
                tried = true;
                if (callback) callback(true);
            };
            
            link.onerror = function() {
                console.warn('Bootstrap: Local path failed:', path);
                if (!tried) {
                    tryNextPath();
                } else if (callback) {
                    callback(false);
                }
            };
            
            if (document.head) {
                document.head.appendChild(link);
            } else {
                var waitForHead = setInterval(function() {
                    if (document.head) {
                        clearInterval(waitForHead);
                        document.head.appendChild(link);
                    }
                }, 50);
                
                // Timeout after 2 seconds
                setTimeout(function() {
                    if (!tried) {
                        clearInterval(waitForHead);
                        tryNextPath();
                    }
                }, 2000);
            }
        }
        
        tryNextPath();
    }
    
    // Load Bootstrap from CDN (only if online)
    function loadBootstrapFallback() {
        try {
            // Check if fallback already loaded
            if (document.getElementById('bootstrap-css-fallback')) {
                return;
            }
            
            // Only try CDN if online
            if (!isOnline()) {
                console.warn('Bootstrap: Offline - cannot load from CDN. Trying alternative local paths.');
                tryLocalBootstrapPaths(function(success) {
                    if (!success) {
                        console.error('Bootstrap: All local paths failed. Bootstrap styles may not be available. The application may not display correctly.');
                    }
                });
                return;
            }
            
            // Create fallback link
            var fallbackLink = document.createElement('link');
            fallbackLink.id = 'bootstrap-css-fallback';
            fallbackLink.rel = 'stylesheet';
            fallbackLink.href = 'https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css';
            
            // Handle CDN load error
            fallbackLink.onerror = function() {
                console.warn('Bootstrap: CDN fallback failed. Trying alternative local paths.');
                tryLocalBootstrapPaths(function(success) {
                    if (!success) {
                        console.error('Bootstrap: All loading methods failed. Bootstrap styles may not be available. The application may not display correctly.');
                    }
                });
            };
            
            if (document.head) {
                document.head.appendChild(fallbackLink);
            } else {
                // Wait for head to be available
                var waitForHead = setInterval(function() {
                    if (document.head) {
                        clearInterval(waitForHead);
                        document.head.appendChild(fallbackLink);
                    }
                }, 50);
            }
        } catch(e) {
            console.error('Error loading Bootstrap fallback:', e);
            // Try local paths as last resort
            tryLocalBootstrapPaths(function(success) {
                if (!success) {
                    console.error('Bootstrap: All fallback methods failed.');
                }
            });
        }
    }
    
    // Check if Bootstrap CSS is actually loaded and working
    function checkBootstrap() {
        try {
            var bootstrapLink = document.getElementById('bootstrap-css');
            if (!bootstrapLink) {
                // Link element doesn't exist
                if (isOnline()) {
                    loadBootstrapFallback();
                }
                return;
            }
            
            // Method 1: Check if stylesheet loaded by checking sheet property
            var sheetLoaded = false;
            try {
                if (bootstrapLink.sheet && bootstrapLink.sheet.cssRules && bootstrapLink.sheet.cssRules.length > 0) {
                    sheetLoaded = true;
                }
            } catch(e) {
                // Cross-origin or CORS issue - this is expected for local files in some browsers
                // The file might still be loaded, we'll use method 2
            }
            
            // Method 2: Check if link is in error state
            var linkError = false;
            if (bootstrapLink.sheet === null && bootstrapLink.href) {
                // Link has href but no sheet - might be loading or failed
                // Wait a bit and check again
                setTimeout(function() {
                    if (bootstrapLink.sheet === null && bootstrapLink.href && 
                        bootstrapLink.href.indexOf('/static/css/bootstrap.min.css') !== -1) {
                        // Still no sheet after delay, try fallback if online
                        if (isOnline()) {
                            loadBootstrapFallback();
                        }
                    }
                }, 1000);
                return; // Exit early, will check again after timeout
            }
            
            // Method 3: Test Bootstrap-specific style (only if body exists)
            var bootstrapLoaded = false;
            if (document.body && document.body.appendChild) {
                try {
                    var testEl = document.createElement('div');
                    testEl.className = 'container';
                    testEl.style.position = 'absolute';
                    testEl.style.visibility = 'hidden';
                    testEl.style.width = '1px';
                    testEl.style.height = '1px';
                    testEl.style.top = '-9999px';
                    testEl.style.left = '-9999px';
                    document.body.appendChild(testEl);
                    
                    var style = window.getComputedStyle(testEl);
                    // Bootstrap container has max-width or padding set
                    bootstrapLoaded = (style.maxWidth && style.maxWidth !== 'none' && style.maxWidth !== 'auto' && style.maxWidth !== '') || 
                                      (style.paddingLeft && parseFloat(style.paddingLeft) > 0) ||
                                      (style.paddingRight && parseFloat(style.paddingRight) > 0);
                    
                    if (document.body.contains(testEl)) {
                        document.body.removeChild(testEl);
                    }
                } catch(e) {
                    // If test fails, assume not loaded (but only if online to try fallback)
                }
            }
            
            // If Bootstrap not detected, use fallback (only if online)
            if (!sheetLoaded && !bootstrapLoaded && !linkError) {
                // Double-check: test if we can see Bootstrap classes
                if (document.body) {
                    var testDiv = document.createElement('div');
                    testDiv.className = 'd-none';
                    testDiv.style.position = 'absolute';
                    testDiv.style.visibility = 'hidden';
                    document.body.appendChild(testDiv);
                    var computed = window.getComputedStyle(testDiv);
                    var hasBootstrap = computed.display === 'none'; // Bootstrap's d-none sets display:none
                    document.body.removeChild(testDiv);
                    
                    if (!hasBootstrap && isOnline()) {
                        loadBootstrapFallback();
                    }
                } else if (isOnline()) {
                    loadBootstrapFallback();
                }
            }
        } catch(e) {
            // If check fails and we're online, try fallback
            if (isOnline()) {
                loadBootstrapFallback();
            }
        }
    }
    
    // Initialize Bootstrap loading check
    function init() {
        try {
            var bootstrapLink = document.getElementById('bootstrap-css');
            if (!bootstrapLink) {
                // Link doesn't exist, load from CDN directly (if online)
                if (isOnline()) {
                    loadBootstrapFallback();
                }
                return;
            }
            
            // Listen for error on link element (will trigger if file doesn't exist)
            bootstrapLink.onerror = function() {
                console.warn('Bootstrap: Primary local file failed to load');
                // Try alternative local paths first
                tryLocalBootstrapPaths(function(success) {
                    if (!success) {
                        // If all local paths fail, try CDN (only if online)
                        loadBootstrapFallback();
                    }
                });
            };
            
            // Check after stylesheet loads
            bootstrapLink.onload = function() {
                setTimeout(checkBootstrap, 100);
            };
            
            // Also check immediately if link seems to have failed
            if (!bootstrapLink.href || bootstrapLink.href === '' || bootstrapLink.href === window.location.href) {
                console.warn('Bootstrap: Link href appears invalid');
                bootstrapLink.onerror();
            }
            
            // Also check after DOM ready with multiple attempts
            function runCheck() {
                setTimeout(checkBootstrap, 300);
                setTimeout(checkBootstrap, 1000); // Second check after 1 second
            }
            
            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', runCheck);
            } else if (document.readyState === 'interactive' || document.readyState === 'complete') {
                runCheck();
            }
            
            // Listen for online/offline events
            window.addEventListener('online', function() {
                // When coming back online, check if Bootstrap loaded
                setTimeout(checkBootstrap, 500);
            });
            
            window.addEventListener('offline', function() {
                // When going offline, make sure we don't try CDN
                var fallback = document.getElementById('bootstrap-css-fallback');
                if (fallback && fallback.href.indexOf('bootstrapcdn.com') !== -1) {
                    // CDN fallback might not work offline, but leave it in case it was cached
                }
            });
            
        } catch(e) {
            console.error('Bootstrap fallback init error:', e);
            // If init fails and we're online, just load fallback
            if (isOnline()) {
                loadBootstrapFallback();
            }
        }
    }
    
    // Start after a small delay to ensure DOM is available
    if (document.head) {
        init();
    } else {
        // Wait for head to be available
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', init);
        } else {
            setTimeout(init, 50);
        }
    }
})();
