let out = document.getElementById('output');
let from_el = document.getElementById('from');
let stat = document.getElementById('status');
let modal = document.getElementById('modal');
let modal_body = document.getElementById('modal-body');
let modal_title = document.getElementById('modal-title');
let modal_close = document.getElementById('modal-close');

modal_close.addEventListener('click', e => { modalClose(); });

window.addEventListener('click', e => {
	if (e.target === modal){ modalClose(); }
});
window.addEventListener('keydown', e => {
	if (e.key === 'Escape'){ modalClose(); }
});

let qtype = document.getElementById('qtype');
qtype.addEventListener('change', e => {
	from_el.value = '0';
	doSearch(0);
});

let size_el = document.getElementById('size');
size_el.addEventListener('change', e => {
	from_el.value = '0';
	doSearch(0);
});

let q = document.getElementById('q');
q.addEventListener('keypress', e => {
	if (e.keyCode == 13){
		doSearch();
		e.preventDefault();
		return false;
	}
});
q.focus();

let qbtn = document.getElementById('qbtn');
qbtn.addEventListener('click', e => { doSearch(0); });

let next = document.getElementById('nextbtn');
next.addEventListener('click', e => { doSearch(1); });

let prev = document.getElementById('prevbtn');
prev.addEventListener('click', e => { doSearch(-1); });

let active = false
function doSearch(dir) {
	if (active) { return; }

	let query = q.value;
	if(query.length < 1){
		next.setAttribute('disabled', 'disabled');
		prev.setAttribute('disabled', 'disabled');
		emptyChildren(stat);
		emptyChildren(out);
		return;
	}
	active = true;

	let stype = Number(qtype.value);
	let from = Number(from_el.value);
	let size = Number(size_el.value);

	if(query !== q.dataset.last){
		from = 0
	} else if(dir === 1) {
		from += size;
	} else if(dir === -1) {
		from = Math.max(from - size, 0);
	}

	let url = '';
	switch (stype) {
		case 1: url = 'genera.json'; break;
		case 2: url = 'species.json'; break;
		case 3: url = 'references.json'; break;
		default: active = false; return;
	}
	url += `?f=${from}&z=${size}&q=${encodeURIComponent(q.value)}`;

	fetch(url).then(response => {
		if (!response.ok) { throw 'response not ok'; }
		return response.json();
	}).then(result => {
		emptyChildren(out);

		result.hits.forEach(r => {
			let pdiv = document.createElement('div');
			pdiv.className = "hit";
			// Stupid javascript tricks: force single values into an array
			for(const s of [].concat(r.source)){
				// Display references differently
				if (stype === 3){
					let div = document.createElement('div');
					div.appendChild(document.createTextNode(s));
					pdiv.appendChild(div);
				} else {
					let a = document.createElement('a');
					a.className = "src";
					a.href = `#${r.id}`
					a.addEventListener('click', e => { doDetail(stype, r.id); });
					a.appendChild(document.createTextNode(s));
					pdiv.appendChild(a);
				}
			}

			if('alt_source' in r){
				for(const s of [].concat(r.alt_source)){
					let div = document.createElement('div');
					div.className = "alt";
					div.appendChild(document.createTextNode(s));
					pdiv.appendChild(div);
				}
			}
			out.appendChild(pdiv);
		});

		emptyChildren(stat);
		if(!out.lastChild){
			out.appendChild(document.createTextNode("No results"));
		} else {
			from_el.value = String(from);
			q.dataset.last = query;
			stat.appendChild(document.createTextNode(
				`Displaying ${from+1} - ${from+result.hits.length} of ${result.total}`
			));
		}

		if((from+result.hits.length) >= result.total){
			next.setAttribute('disabled', 'disabled');
		} else { next.removeAttribute('disabled'); }

		if(from === 0){ prev.setAttribute('disabled', 'disabled'); }
		else { prev.removeAttribute('disabled'); }

		active = false;
	}).catch(err => {
		if(window.console){ console.log(err); }
		active = false;
	});
}

function doDetail(stype, id){
	if (active) { return; }
	active = true;

	document.activeElement.blur();

	let url = '';
	switch(stype){
		case 1: url = 'genera_full.json'; break;
		case 2: url = 'species_full.json'; break;
		default: active = false; return;
	}

	fetch(`${url}?id=${id}`).then(response => {
		if (!response.ok) { throw 'response not ok'; }
		return response.json();
	}).then(r => {
		emptyChildren(modal_title);
		switch(stype){
			case 1:
				modal_title.appendChild(document.createTextNode('Genera Detail'));
			break;
			case 2:
				modal_title.appendChild(document.createTextNode('Species Detail'));
			break;
		}

		emptyChildren(modal_body);
		for(const s of [].concat(r.source)){
			let div = document.createElement('div');
			div.className = "src";
			div.appendChild(document.createTextNode(s));
			modal_body.appendChild(div);
		}

		if('alt_source' in r){
			for(const s of [].concat(r.alt_source)){
				let div = document.createElement('div');
				div.className = "alt";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('comment' in r){
			for(const s of [].concat(r.comment)){
				let div = document.createElement('div');
				div.className = "cmt";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('species_source' in r){
			for(const s of [].concat(r.species_source)){
				let div = document.createElement('div');
				div.className = "spc";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('occurance' in r){
			for(const s of [].concat(r.occurance)){
				let div = document.createElement('div');
				div.className = "occ";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		modal.style.display = 'block';
		active = false;
	}).catch(err => {
		if(window.console){ console.log(err); }
		active = false;
	});
}

function emptyChildren(el){ while(el.lastChild) el.removeChild(el.lastChild); }

function modalClose(){
	if (modal.style.display === 'block'){
		modal.scrollTop = 0;
		modal.style.display = 'none';
		q.focus();
	}
}
