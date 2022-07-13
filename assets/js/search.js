let out = document.getElementById('output');
let from_el = document.getElementById('from');
let stat = document.getElementById('status');
let modal = document.getElementById('modal');
let modal_body = document.getElementById('modal-body');
let modal_title = document.getElementById('modal-title');
let modal_close = document.getElementById('modal-close');

window.addEventListener('click', e => {
	if (e.target === modal) closeModal();
});
window.addEventListener('keydown', e => {
	if (e.key === 'Escape') closeModal();
});

let qtype = document.getElementById('qtype');

let size_el = document.getElementById('size');

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
qbtn.addEventListener('click', e => { doSearch(); });

let next = document.getElementById('nextbtn');
next.addEventListener('click', e => { doSearch(1); });

let prev = document.getElementById('prevbtn');
prev.addEventListener('click', e => { doSearch(-1); });

function closeModal(){
	if (modal.style.display === 'block'){
		modal.scrollTop = 0;
		document.body.style.overflow = '';
		modal.style.display = 'none';
		saveHash();
	}
}

modal_close.addEventListener('click', closeModal);

let sactive = false;
function doSearch(dir,noupdate) {
	if (sactive) { return; }

	let query = q.value;
	if(query.length < 1){
		next.setAttribute('disabled', 'disabled');
		prev.setAttribute('disabled', 'disabled');
		emptyChildren(stat);
		emptyChildren(out);
		return;
	}
	sactive = true;

	let stype = Number(qtype.value);
	let from = Number(from_el.value);
	let size = Number(size_el.value);

	if(q.dataset.last && query !== q.dataset.last){
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
		default: sactive = false; return;
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
					a.className = "src ndt";
					a.href = `#${r.id}`
					a.addEventListener('click', e => {
						doDetail(stype, r.id);
						e.preventDefault();
						return false;
					});
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
			stat.appendChild(document.createTextNode(
				`Displaying ${from+1} - ${from+result.hits.length} of ${result.total}`
			));
		}
		q.dataset.last = query;

		if((from+result.hits.length) >= result.total){
			next.setAttribute('disabled', 'disabled');
		} else { next.removeAttribute('disabled'); }

		if(from === 0){ prev.setAttribute('disabled', 'disabled'); }
		else { prev.removeAttribute('disabled'); }

		if (!noupdate) saveHash();
		sactive = false;
	}).catch(err => {
		if(window.console) console.log(err);
		sactive = false;
		alert('An error occurred while talking to server, please try again.');
	});
}

let dactive = false;
function doDetail(stype, id){
	if (dactive) { return; }
	dactive = true;

	document.activeElement.blur();

	let url = '';
	switch(stype){
		case 1: url = 'genera_full.json'; break;
		case 2: url = 'species_full.json'; break;
		default: dactive = false; return;
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

		let h3 = document.createElement('h3');
		h3.appendChild(document.createTextNode('Preferred Name'));
		modal_body.appendChild(h3);
		for(const s of [].concat(r.source)){
			let div = document.createElement('div');
			div.className = "src";
			div.appendChild(document.createTextNode(s));
			modal_body.appendChild(div);
		}

		if('alt_source' in r){
			let h3 = document.createElement('h3');
			h3.appendChild(document.createTextNode('Synonymies'));
			modal_body.appendChild(h3);

			for(const s of [].concat(r.alt_source)){
				let div = document.createElement('div');
				div.className = "alt";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('comment' in r){
			let h3 = document.createElement('h3');
			h3.appendChild(document.createTextNode('Comments'));
			modal_body.appendChild(h3);

			for(const s of [].concat(r.comment)){
				let div = document.createElement('div');
				div.className = "cmt";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('species_source' in r){
			let h3 = document.createElement('h3');
			h3.appendChild(document.createTextNode('Species'));
			modal_body.appendChild(h3);

			for(const s of [].concat(r.species_source)){
				let div = document.createElement('div');
				div.className = "spc";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('occurance' in r){
			let h3 = document.createElement('h3');
			h3.appendChild(document.createTextNode('Occurances'));
			modal_body.appendChild(h3);

			for(const s of [].concat(r.occurance)){
				let div = document.createElement('div');
				div.className = "occ";
				div.appendChild(document.createTextNode(s));
				modal_body.appendChild(div);
			}
		}

		if('images' in r){
			let h3 = document.createElement('h3');
			h3.appendChild(document.createTextNode('Images'));
			modal_body.appendChild(h3);

			let div = document.createElement('div');
			div.className = 'img';
			for(const s of [].concat(r.images)){
				let img = document.createElement('img')
				img.src = `images/${s}`;
				div.appendChild(img);
			}
			modal_body.appendChild(div)
		}

		modal.style.setProperty('display', 'block');
		modal.style.display = 'block';
		document.body.style.overflow = 'hidden';
		saveHash(id);
		dactive = false;
	}).catch(err => {
		if(window.console) console.log(err);
		dactive = false;
		alert('An error occurred while talking to server, please try again.');
	});
}

function emptyChildren(el){
	while(el.lastChild) el.removeChild(el.lastChild);
}

function saveHash(id) {
	let oldhash = decodeURIComponent(location.hash.substring(1));
	let newhash = `${qtype.value},${JSON.stringify(q.value)},${size_el.value},${from_el.value}`;
	if(id) newhash += `,${JSON.stringify(id)}`;
	if (newhash != oldhash) {
		history.pushState({}, '', `${location.pathname}#${newhash}`);
	}
}

// On demand, check if there's a hash, if there is
// load it as a search, and if there's a detail modal id,
// load the modal, too.
function loadHash(){
	if(location.hash){
		try {
			let jsn = decodeURIComponent(location.hash.substring(1));
			let arr = JSON.parse(`[${jsn}]`);
			qtype.value = arr[0];
			q.value = arr[1];
			size_el.value = arr[2];
			from_el.value = arr[3];

			doSearch(0, true);
			if (arr.length > 4){ doDetail(arr[0], arr[4], true); }
			else { closeModal(); }
		} catch(err) {
			if(window.console) console.log(err);
		}
	}
}

loadHash();

window.addEventListener('hashchange', loadHash);

// These need to be initialized last so they don't interfere
// with the hash loading
qtype.addEventListener('change', e => {
	from_el.value = '0';
	doSearch();
});

size_el.addEventListener('change', e => {
	from_el.value = '0';
	doSearch();
});
