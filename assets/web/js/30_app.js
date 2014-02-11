$(document).ready(function(){
    initMenu(Menu);
    Menu.Button.click();
});

var Menu = {
    Button: $("#fileMenu"),
    List: $("#fileList"),
    Active: false
};

function initMenu(m) {
    m.Button.on("click", toggleMenu(m));
    // m.Button.addEventListener("click", toggleMenu(m), false);
}

function toggleMenu(m) {
    return function() {
	p = m.Button.position();
	m.List.offset({
	    left: p.left,
	    top: p.top + 41});
    }
}
