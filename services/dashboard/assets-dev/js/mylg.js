$( document ).ready(function() {
        document.querySelector('.mdl-layout__drawer').addEventListener('click', function () {
            document.querySelector('.mdl-layout__obfuscator').classList.remove('is-visible');
            this.classList.remove('is-visible');
        }, false);
});


