// include file for index.php  :)
var current_view='';		//  category or feed - for what's currently viewed
var current_view_id='';		//  id of the category or feed currently being viewed.
var current_entry_id='';	//  id of current entry being viewed
var status_div='left_notify'; //  id of the status div
//Stuff for the arrow navigation, from http://api.jquery.com/keydown/
$(window).keydown(function (e) {
		// test for in an input box so as not to forward when moving around
		if( document.activeElement instanceof HTMLInputElement ) 
		{
			return;
		}
		if(e.which ==85 || e.which == 68) // u,d
		{
			toggle_visible('entries_list_div');
			return;
		}
		if(e.which == 77) // m
		{
			toggleMark(current_entry_id);
		}
		if(e.which==39 || e.which == 75 || e.which == 32) // ->, k, (space)
		{
	  		e.preventDefault();
			showNextEntry(current_entry_id);
		}
		else if(e.which==37 || e.which == 74) // <-, j
		{
			e.preventDefault();
			showPreviousEntry(current_entry_id);
		}
	});

$(function() {
    $('form').each(function() {
        $(this).find('input').keypress(function(e) {
            // Enter pressed?
            if(e.which == 10 || e.which == 13) {
                this.form.submit();
            }
        });

        $(this).find('input[type=submit]').hide();
    });
});

function showPreviousEntry(id) {
	vals = PrevNextTable(id);
	if(vals==null){return;}
	show_entry(vals.prev);
}
function showNextEntry(id) {
	vals = PrevNextTable(id);
	if(vals==null){return;}
	show_entry(vals.next);
}
function PrevNextTable(id) {
	var prev=0;
	var curr=0;
	var next=0;
	var table=document.getElementById('headlinesList');
	var rowLength= entries_length();
	var prevrow = null;
	var nextrow = null;
	for(var i=0;i<rowLength;i++) {
		var row=table.rows[i];
		if(id=="") {
			nextrow=row.id;
			prevrow=table.rows[rowLength-1].id;
			break;
		}
		if(row.id== "RROW-"+id) {
			if(table.rows[i-1] != null) { prevrow=table.rows[i-1].id;}
			if(table.rows[i+1] != null) { nextrow=table.rows[i+1].id;}
			curr=row.id.replace("RROW-","");
			break;
		}
	}
	if(prevrow != null) {prev=prevrow.replace("RROW-","");}
	if(nextrow != null) {next=nextrow.replace("RROW-","");}
	return{prev:prev,curr:curr,next:next}
}
// Update the link for a given feed, this isn't update because it needs to POST the url
function update_link(fc,id,form){
	document.getElementById('menu_status').innerHTML='Updating...';
	var link        =form.update_link_text.value;
	link            =encodeURIComponent(link);
	url                     ="url="+link
	$.ajax({type: "POST",url: '/'+fc+'/'+id+'/link/', data: url,success:function(html){$('#menu_status').html(html);}})
}
function update(fc, id, todo, form) {
	try{val = encodeURIComponent(form.val.value);}catch(err){}
	$.ajax({type: "GET",url: fc+"/"+id+"/"+todo+"/"+val, success:function(html){
		$('#menu_status').html(html);
	}})
}

function update_entries(fc, id) {
    $.ajax({type: "GET", url: fc+"/"+id+"/update/true",  success:function(html){
        $('#menu_status').html(html);
        update_count(fc,id);
    }})
}
//toggle the visibility of a given passed div id
function toggle_visible(id) {
	$('#'+id).toggle();
	set_entryview();
}

// Hides the table row of a given id, and makes an ajax call to mark read in the backend.
function remove_entry(id) {
	oldEl = document.getElementById('RROW-'+id);
	parentEl = oldEl.parentNode;
	parentEl.removeChild(oldEl);	
	$.ajax({type: "GET",url: '/entry/mark/_/_/'+id+'/read', success:function(html){$('#status_div').html(html);}})
	if(oldEl.className.match(/unread/) == 'unread'){decrement_count();}
}
// Toggles the marked/unmarked for a given id, and replaces the div with the src for the appropriate image.
function toggleMark(id)  {
	var mark_div='FMARKPIC-' +id;
	var page_mark_div='EMARKPIC-' +id;
	$.ajax({type: "GET",url: '/entry/mark/ / /'+id+'/togglemarked', success:function(html){
		try{document.getElementById('EMARKPIC-'+id).innerHTML=html;}catch(err){}
		try{document.getElementById('FMARKPIC-'+id).innerHTML=html;}catch(err){}
	}})
}
function set_unread_count(fc,id,ct) {
    var name_div=fc=='feed'?"FEEDROW-"+id:"CATROW-"+id;
    var oddness = ct >=1 ? 'oddUnread' : 'odd';
    try{document.getElementById('FEEDU-'+id).innerHTML=ct;}catch (err){}
    try{document.getElementById(name_div).className=oddness;}catch (err){}
}
function entries_length() {
    var table = document.getElementById('headlinesList')
    return table.rows.length;
}
function update_count(fc,id) {
	if(fc == 'feed')
		{name_div='FEEDROW-'+current_view_id;}
	else
		{name_div='CATROW-'+current_view_id;}
	var unread_div='FEEDU-'+id;
	url="/"+fc+"/"+id+"/unread";
	$.ajax({type: "GET", url: url, success:function(ct){
		eo = ct < 1? 'odd' : 'evenUnread';
		try{document.getElementById(unread_div).innerHTML=ct;}catch (err){}
		try{document.getElementById(name_div).className=eo;}catch (err){}
	}})
}
function all_entry_ids() {
    var table=document.getElementById('headlinesList');
    var all_entries = new Array;
    for(var i=0;i<table.rows.length;i++) {
        var row=table.rows[i];
        var ri=row.id.replace("RROW-","");
        all_entries.push(ri);
    }
    return all_entries;
}
// Lowers the unread count for the current_view_id by one.
function decrement_count(dc) {
    dc= dc || 1;
	var cr_view_div='FEEDU-'+current_view_id;
	var current_value=document.getElementById(cr_view_div).innerHTML;
	current_value=current_value-dc;
	if (current_value < 1) { current_value=0;}
    set_unread_count(current_view,current_view_id,current_value);
}
//Set the height of the content and entry divs 
function set_entryview() {
	// If the entries list is hidden, we have to change the height or it's irritating
	if($('#entries_list_div').is(":hidden")){
		$('#content_container').css("height","95%");
		$('#entry_content').css("height","95%");
	} else {
		$('#content_container').css("height","70%");
		$('#entry_content').css("height","70%");
	}
}
// Populates view_div with the content for a given id.
function show_entry(id) {
	list_row=document.getElementById('RROW-'+id);
	current_entry_id=id;
	$.ajax({type: "GET",url: "/entry/"+id, success:function(html){
		$('#view_div').html(html);
		set_entryview();
	}})
	scrollup('view_div');
	try{if(list_row.className.match(/unread/) == 'unread'){decrement_count();}}catch(err){}
	try{list_row.className=list_row.className.replace("unread","");}catch(err){}
}
// Populates the feeds_div with a list of feeds.
function feedList() {
	document.getElementById('feeds_status').innerHTML='<img src="static/mozilla_giallo.gif" height="10">';
	current_view='feed';
	$.ajax({type: "GET",url: '/feed/list/', success:function(html){$('#feeds_div').html(html);document.getElementById('feeds_status').innerHTML='';}})
}

function mark_list_read(fc, id) {
    var ids=all_entry_ids();
    while (ids.length > 0) {
        var temparr = ids.slice(0,500);
        ids=ids.slice(500);
        data=temparr.join();
        decrement_count(ids.length);
        $.ajax({
            type: "GET",
            url: '/entry/mark/'+fc+'/'+id+'/'+data+'/read',
            success:function(html){
                $('#entries_list_div').html(html);
                $('#menu_status').html('');
                set_unread_count(fc,id, 0);
            }
        });
    }
}
// Populates the feeds_div with a list of categories.
// If id is given, shows the feeds inside that category
function categoryList(id) {
	document.getElementById('feeds_status').innerHTML='<img src="static/mozilla_giallo.gif" height="10">';
	current_view='category';
	$.ajax({type: "GET",url: "/categoryList/"+id, success:function(html){$('#feeds_div').html(html);	document.getElementById('feeds_status').innerHTML='';
}})
}
function search(feedcat,id,form) {
	try{val = form.val.value;}catch(err){alert(err);}
	var index=form.search_select.selectedIndex;
    var selvalue=form.search_select.options[index].value;
	if( val == "" ) {
		entries(feedcat,id,selvalue);
		return;
	}
	path=feedcat+"/"+id+"/search/"+val+"/"+selvalue;
	try{document.getElementById('menu_status').innerHTML='Loading...';}catch(err){} // May be null
	$.ajax({
		type: "GET",
		url: '/entries/'+path, 
		success: function(html){
			$('#entries_list_div').html(html);
			document.getElementById('menu_status').innerHTML='';
			if($('#entries_list_div').is(":hidden")){
				$('#entries_list_div').toggle();
			}
			scrollup('entries_list_div');
		}
	});
}
function entries(feedcat,id,mode) {
	current_view=feedcat;
	current_view_id=id;
	path=feedcat+"/"+id+"/"+mode;
    if(feedcat == 'feed')
        {name_div='FEEDROW-'+current_view_id;}
    else
        {name_div='CATROW-'+current_view_id;}
	try{document.getElementById('menu_status').innerHTML='Loading...';}catch(err){} // May be null
	$.ajax({type: "GET",url: '/menu/'+path, success:function(html){$('#settings_div').html(html);}})
	$.ajax({
		type: "GET",
		url: '/entries/'+path, 
		success: function(html){
			$('#entries_list_div').html(html);
			document.getElementById('menu_status').innerHTML='';
			if($('#entries_list_div').is(":hidden")){
				$('#entries_list_div').toggle();
			}
			scrollup('entries_list_div');
			current_entry_id='';
			if(mode=="unread") {
			    set_unread_count(feedcat,id,entries_length());
			}
		}
	});
}
function customize(form) {
	document.getElementById('menu_status').innerHTML='Loading...';
	var index=form.select.selectedIndex;
	var selvalue=form.select.options[index].value;
	if(selvalue == ''){return;}
	if(selvalue == 'delete'){url='/feed/'+current_view_id+'/delete/';}
	else if(selvalue=='default' || selvalue=='link' || selvalue=='extended'||selvalue=='proxy')
		{
		url='/feed/'+current_view_id+'/view_mode/'+selvalue
		}
	else
		{
		url='/feed/'+current_view_id+'/category/'+selvalue
		}
	$.ajax({type: "GET",url: url,success:function(html)
	{
		$('#menu_status').html(html);
		$.ajax({type: "GET",url: '/menu/select/'+current_view_id,success:function(html){$('#customize_dropdown').html(html);}})
	}})
}
function add_category(form) {
	$('menu_status').innerHTML='Adding...';
	var newcat =form.add_category_text.value;
	$.ajax({type: "GET",url: '/category/ /new/'+newcat,success:function(html)
	{
		$('#menu_status').html(html);
		form.add_category_text.value="";
		if(current_view == 'category') { categoryList(); }
	}})
}
function add_feed(form) {
	$('menu_status').innerHTML='Adding...';
	var newfeed	=form.add_feed_text.value;
	newfeed		=encodeURIComponent(newfeed);
	url			="url="+newfeed;
	$.ajax({type: "POST",url: '/feed/new/', data:url,success:function(html)
	{
		$('#menu_status').html(html);
		form.add_feed_text.value="";
		if(current_view == 'category') { categoryList(); }
		if(current_view == 'feed') {feedList();}
	}})
}
//Just for scrolling to the top when loading something
function scrollup(id) {
	try{document.getElementById(id).scrollTop=0;}catch(err){}
}
function scrollto(id,to) {
	try{document.getElementById(id).scrollTop=to;}catch(err){}
}
