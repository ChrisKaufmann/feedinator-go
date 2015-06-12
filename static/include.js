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
		if(e.which == 38 || e.which == 85 || e.which ==40 || e.which == 68) // up,down,u,d
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

function showPreviousEntry(id)
{
	document.getElementById('menu_status').innerHTML='Getting previous id';
	$.ajax({type: "GET",url: '/entries/'+current_view+"/"+current_view_id+"/previous/"+id, success:function(html){
		if(html == "0"){
			$('#menu_status').html("No previous entries");
			return;
		}
		$('#menu_status').html(html);
		show_entry(html);
	}});
}
function showNextEntry(id)
{
	document.getElementById('menu_status').innerHTML='Getting next id';
	$.ajax({type: "GET",url: "/entries/"+current_view+"/"+current_view_id+"/next/"+id, success:function(html){
		if(html == "0"){
			$('#menu_status').html("No new entries");
			return;
		}
		$('#menu_status').html(html);
		show_entry(html);
	}});
}
// Update the link for a given feed, this isn't update because it needs to POST the url
function update_link(fc,id,form)
{
	document.getElementById('menu_status').innerHTML='Updating...';
	var link        =form.update_link_text.value;
	link            =encodeURIComponent(link);
	url                     ="url="+link
	$.ajax({type: "POST",url: '/'+fc+'/'+id+'/link/', data: url,success:function(html){$('#menu_status').html(html);}})
}
function update(fc,id,todo,form)
{
	document.getElementById('menu_status').innerHTML='Updating...';
	try{val = form.val.value;}catch(err){}
	try{val = encodeURIComponent(val);}catch(err){}
	$.ajax({type: "GET",url: fc+"/"+id+"/"+todo+"/"+val, success:function(html){
		$('#menu_status').html(html);
		if(todo == "update"){update_count(fc,id);}
	}})
}

//toggle the visibility of a given passed div id
function toggle_visible(id)
{
	$('#'+id).toggle();
	set_entryview();
}

// Hides the table row of a given id, and makes an ajax call to mark read in the backend.
function remove_entry(id)
{
	oldEl = document.getElementById('RROW-'+id);
	parentEl = oldEl.parentNode;
	parentEl.removeChild(oldEl);	
	$.ajax({type: "GET",url: '/entry/mark/'+id+'/read', success:function(html){$('#status_div').html(html);}})
	if(oldEl.className.match(/unread/) == 'unread'){decrement_count();}
}
// Toggles the marked/unmarked for a given id, and replaces the div with the src for the appropriate image.
function toggleMark(id) 
{
	var mark_div='FMARKPIC-' +id;
	var page_mark_div='EMARKPIC-' +id;
	$.ajax({type: "GET",url: '/entry/mark/ / /'+id+'/togglemarked', success:function(html){
		try{document.getElementById('EMARKPIC-'+id).innerHTML=html;}catch(err){}
		try{document.getElementById('FMARKPIC-'+id).innerHTML=html;}catch(err){}
	}})
}
function reportError(request) 
{
	alert("There was a problem");
}
// Set the current_view_id to look read, and the number unread to zero.  
// Could throw errors if the div has since been hidden or removed.
function empty_count()
{
	if(current_view == 'feed')
		{name_div='FEEDROW-'+current_view_id;}
	else
		{name_div='CATROW-'+current_view_id;}
	var unread_div='FEEDU-'+current_view_id;
	// These are expected errors and should not print errors.  
	// expected because the div could have been hidden.
	try{document.getElementById(unread_div).innerHTML='0';}catch (err){} 
	try{document.getElementById(name_div).className='odd';}catch (err){}
}
function update_count(fc,id)
{
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
// Lowers the unread count for the current_view_id by one - if that zeroes, calls empty_count.
function decrement_count()
{
	var cr_view_div='FEEDU-'+current_view_id;
	var current_value=document.getElementById(cr_view_div).innerHTML;
	current_value=current_value-1;
	document.getElementById(cr_view_div).innerHTML=current_value;
	if(current_value<=0){empty_count();}
}
//Set the height of the content and entry divs 
function set_entryview()
{
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
function show_entry(id)
{
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
function feedList()
{
	document.getElementById('feeds_status').innerHTML='<img src="static/mozilla_giallo.gif" height="10">';
	current_view='feed';
	$.ajax({type: "GET",url: '/feed/list/', success:function(html){$('#feeds_div').html(html);document.getElementById('feeds_status').innerHTML='';}})
}

function mark_list_read(fc, id)
{
	data=$("#entries_form").serialize();
	ndata=data.replace(/id%5B%5D=/g,',')
	data=ndata.replace(/&/g,'')
	$.ajax({type: "GET",url: '/entry/mark/'+fc+'/'+id+'/'+data+'/read', success:function(html){$('#entries_list_div').html(html);$('#menu_status').html('');empty_count();}});
}
// Populates the feeds_div with a list of categories.
// If id is given, shows the feeds inside that category
function categoryList(id)
{
	document.getElementById('feeds_status').innerHTML='<img src="static/mozilla_giallo.gif" height="10">';
	current_view='category';
	$.ajax({type: "GET",url: "/categoryList/"+id, success:function(html){$('#feeds_div').html(html);	document.getElementById('feeds_status').innerHTML='';
}})
}
function search(feedcat,id,form)
{
	try{val = form.val.value;}catch(err){alert(err);}
	var index=form.search_select.selectedIndex;
    var selvalue=form.search_select.options[index].value;
	if( val == "" ) {
		entries(feedcat,id,selvalue);
		return;
	}
	path=feedcat+"/"+id+"/search/"+val+"/"+selvalue;
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
		}
	});
}
function entries(feedcat,id,mode)
{
	current_view=feedcat;
	current_view_id=id;
	path=feedcat+"/"+id+"/"+mode;
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
		}
	});
}
function customize(form)
{
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
function add_category(form)
{
	$('menu_status').innerHTML='Adding...';
	var newcat =form.add_category_text.value;
	$.ajax({type: "GET",url: '/category/ /new/'+newcat,success:function(html)
	{
		$('#menu_status').html(html);
		form.add_category_text.value="";
		if(current_view == 'category') { categoryList(); }
	}})
}
function add_feed(form)
{
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
function scrollup(id)
{
	try{document.getElementById(id).scrollTop=0;}catch(err){}
}
function scrollto(id,to)
{
	try{document.getElementById(id).scrollTop=to;}catch(err){}
}
