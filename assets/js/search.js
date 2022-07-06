let out = document.getElementById('output');
let from_el = document.getElementById('from');
let stat = document.getElementById('status');

let qtype = document.getElementById('qtype');
qtype.addEventListener('change', function(){
	from_el.value = '0';
	doSearch(0);
});

let size_el = document.getElementById('size');
size_el.addEventListener('change', function(){
	from_el.value = '0';
	doSearch(0);
});

let q = document.getElementById('q');
q.addEventListener('keypress', function(event){
	if (event.keyCode == 13){
		doSearch();
		event.preventDefault();
		return false;
	}
});
q.focus();

let qbtn = document.getElementById('qbtn');
qbtn.addEventListener('click', function(){ doSearch(0); });

let next = document.getElementById('nextbtn');
next.addEventListener('click', function(){ doSearch(1); });

let prev = document.getElementById('prevbtn');
prev.addEventListener('click', function(){ doSearch(-1); });

let active = false
function doSearch(dir) {
	if (active) { return; }

	let query = q.value;
	if(query.length < 1){
		next.setAttribute('disabled', 'disabled');
		prev.setAttribute('disabled', 'disabled');
		while(stat.lastChild) stat.removeChild(stat.lastChild);
		while(out.lastChild) out.removeChild(out.lastChild);
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
		while(out.lastChild) out.removeChild(out.lastChild);

		result.hits.forEach(r => {
			let pdiv = document.createElement('div');
			pdiv.className = "hit";
			// Stupid javascript tricks: force single values into an array
			for(const s of [].concat(r.source)){
				// Display references differently
				if (stype === 3){
					let div = document.createElement('a');
					div.className = "src";
					div.appendChild(document.createTextNode(s));
					pdiv.appendChild(div);
				} else {
					let a = document.createElement('a');
					a.className = "src";
					a.href = '#';
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

		while(stat.lastChild) stat.removeChild(stat.lastChild);
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
