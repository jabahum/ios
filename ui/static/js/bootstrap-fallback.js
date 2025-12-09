// Bootstrap CSS fallback - check if local file loaded, fallback to CDN if not
(function() {
    'use strict';
    
    function loadBootstrapFallback() {
        try {
            // Check if fallback already loaded
            if (document.getElementById('bootstrap-css-fallback')) {
                return;
            }
            
            // Create fallback link
            var fallbackLink = document.createElement('link');
            fallbackLink.id = 'bootstrap-css-fallback';
            fallbackLink.rel = 'stylesheet';
            fallbackLink.href = 'https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css';
            if (document.head) {
                document.head.appendChild(fallbackLink);
            }
        } catch(e) {
            // Silently fail if there's an error
            console.error('Error loading Bootstrap fallback:', e);
        }
    }
    
    function checkBootstrap() {
        try {
            var bootstrapLink = document.getElementById('bootstrap-css');
            if (!bootstrapLink) {
                loadBootstrapFallback();
                return;
            }
            
            // Method 1: Check if stylesheet loaded by checking sheet property
            var sheetLoaded = false;
            try {
                if (bootstrapLink.sheet && bootstrapLink.sheet.cssRules && bootstrapLink.sheet.cssRules.length > 0) {
                    sheetLoaded = true;
                }
            } catch(e) {
                // Cross-origin or not loaded - this is expected for local files
                // Try method 2 instead
            }
            
            // Method 2: Test Bootstrap-specific style (only if body exists)
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
                    document.body.appendChild(testEl);
                    
                    var style = window.getComputedStyle(testEl);
                    // Bootstrap container has max-width and padding
                    bootstrapLoaded = (style.maxWidth && style.maxWidth !== 'none' && style.maxWidth !== 'auto') || 
                                      (style.paddingLeft && parseFloat(style.paddingLeft) > 0) ||
                                      (style.paddingRight && parseFloat(style.paddingRight) > 0);
                    
                    if (document.body.contains(testEl)) {
                        document.body.removeChild(testEl);
                    }
                } catch(e) {
                    // If test fails, assume not loaded
                }
            }
            
            // If Bootstrap not detected, use fallback
            if (!sheetLoaded && !bootstrapLoaded) {
                loadBootstrapFallback();
            }
        } catch(e) {
            // Silently fail and load fallback
            loadBootstrapFallback();
        }
    }
    
    // Wait for DOM to be ready
    function init() {
        try {
            var bootstrapLink = document.getElementById('bootstrap-css');
            if (!bootstrapLink) {
                // Link doesn't exist, load from CDN directly
                loadBootstrapFallback();
                return;
            }
            
            // Listen for error on link element
            bootstrapLink.onerror = function() {
                loadBootstrapFallback();
            };
            
            // Check after stylesheet loads
            bootstrapLink.onload = function() {
                setTimeout(checkBootstrap, 100);
            };
            
            // Also check after DOM ready
            if (document.readyState === 'loading') {
                document.addEventListener('DOMContentLoaded', function() {
                    setTimeout(checkBootstrap, 500);
                });
            } else if (document.readyState === 'interactive' || document.readyState === 'complete') {
                setTimeout(checkBootstrap, 500);
            }
        } catch(e) {
            // If init fails, just load fallback
            loadBootstrapFallback();
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

