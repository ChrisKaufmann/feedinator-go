// include file for index.php  :)
var backend='';// 'http://www.feedinator.com/gui/backend.php';
var current_view='';		//  category or feed - for what's currently viewed
var current_view_id='';		//  id of the category or feed currently being viewed.
var current_entry_id='';	//  id of current entry being viewed
var set_mark_id='';		//  id of an entry to toggle the mark - normally not used.
var status_div='left_notify'; //  id of the status div
var entries_data='';		// hash containing all data about the entries
//Stuff for the arrow navigation, from http://api.jquery.com/keydown/
	$(window).keydown(function (e) {
		if(e.which==39)
		{
	  		e.preventDefault();
			showNextEntry(current_entry_id);
		}
		else if(e.which==37)
		{
			e.preventDefault();
			showPreviousEntry(current_entry_id);
		}
	});
// Take a json array and populate the entries_list_div
function populate_list()
{
	var json=entries_data;
	var div='entries_list_div';
	$(div).innerHTML='';
	var table_data='';
	scrollup(div);
	table_data='<table class="headlinesList" id="headlinesList" width="100%">';
	// do populate stuff here
	for(i in json.entries)
	{
		var readunread=json.entries[i].unread?'unread':'read';
		var link='javascript:show_entry('+i+');';
		table_data+="<tr class='even"+readunread+"' id='RROW-"+i+"'>";
		table_data+='<td><div id="FMARKPIC-'+i+'"><img src="images/'+json.entries[i].img+'" onclick="javascript:toggleMark('+i+');"></div></td>';
		table_data+="<td><a href='javascript:remove_entry("+i+");'>-</a></td>";
		table_data+="<td width='22%'><a href='"+link+"'>"+json.entries[i].feed_name+"</a></td>";
		table_data+="<td width='60%'><a href='"+link+"'>"+json.entries[i].title+"</a></td>";
		table_data+="<td width='12%'>"+json.entries[i].updated+"</td>";
		table_data+="<td><a href='"+json.entries[i].link+"' target='_blank'>-></a></td>";
		table_data+='</tr>\n';
	}
	table_data+="</table>";
	$(div).innerHTML+=table_data;
}

function showPreviousEntry(id)
{
	var url='op=previous_entry&id='+id+'&view_mode='+current_view+'&view_mode_id='+current_view_id;
	document.getElementById('menu_status').innerHTML='Getting next id';
	$.ajax({type: "GET",url: '/entries/'+current_view+"/"+id+"/previous/"+current_view_id, success:function(html){
		$('#menu_status').html(html);
		show_entry(html);
	}});
}
function showNextEntry(id)
{
	document.getElementById('menu_status').innerHTML='Getting next id';
	$.ajax({type: "GET",url: "/entries/"+current_view+"/"+id+"/next/"+current_view_id, success:function(html){
		$('#menu_status').html(html);
		show_entry(html);
	}});
}

// Update the link for a given feed
function update_link(form)
{
	document.getElementById('menu_status').innerHTML='Updating...';
	var link	=form.update_link_text.value;
	link		=encodeURIComponent(link);
	var url		='op=update_feed_link&id='+eurrent_view_id+'&skip_div=1&link='+link;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
}

//update the exclude list for a given feed
function update_exclude(form)
{
	document.getElementById('menu_status').innerHTML='Updating...';
	var list	=form.rename_exclude_text.value;
	var url		='op=update_exclude_list&id='+current_view_id+'&exclude_list='+list;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
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
	set_mark_id=id;
	var mark_div='FMARKPIC-' +id;
	var page_mark_div='EMARKPIC-' +id;
	$.ajax({type: "GET",url: '/entry/mark/'+id+'/togglemarked', success:function(html){
		try{document.getElementById('EMARKPIC-'+id).innerHTML=html;}catch(err){}
		try{document.getElementById('FMARKPIC-'+id).innerHTML=html;}catch(err){}
	}})
}
function reportError(request) 
{
	alert("There was a problem");
}
// Populate the list_div with the entries for a feed of the given id
function feed_entries(id)
{
	try{document.getElementById('menu_status').innerHTML='Loading...';}catch(err){} // may be null
	current_view='feed';
	current_view_id=id;
	var div='entries_list_div';
	var menuurl='op=print_menu&id='+id+'&view_mode=feed';
	scrollup(div)
	$.ajax({type: "GET",url: "menu/feed/"+id, success:function(html){$('#settings_div').html(html);}})
	$.ajax({type: "GET",url: "entries/feed/"+id+"/unread", success:function(html){
		$('#entries_list_div').html(html);
		if($('#entries_list_div').is(":hidden")){
			$('#entries_list_div').toggle();
        }
	}})

	try{$('#menu_status').text='';}catch(err){} // may be null
}
//populate the list_div with read entries
function view_read()
{
	try{document.getElementById('menu_status').innerHTML='Loading...';}catch(err){} // may be nul
	var url='op='+current_view+'_entries&id=' + current_view_id+'&view_read=1';
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#entries_list_div').html(html);}})
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
	try{entries_data=[]}catch (err){}
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
//populate list_div with the entries for a given category id
function category_entries(id)
{
	try{document.getElementById('menu_status').innerHTML='Loading...';}catch(err){}	// May be null
	current_view='category';
	var div='entries_list_div';
	current_view_id=id;
	var menuurl='op=print_menu&view_mode=category&id=' +id;
	scrollup(div);
//	$.ajax({type: "GET",url: '/categoryList', success:function(html){$('#settings_div').html(html);}})
	$.ajax({type: "GET",url: '/menu/category/'+id, success:function(html){$('#settings_div').html(html);}})
	$.ajax({type: "GET",url: "/entries/category/"+id+"/unread", success:function(html){
		$('#entries_list_div').html(html);
		if($('#entries_list_div').is(":hidden")){
			$('#entries_list_div').toggle();
		}
	}})
}
// Populates view_div with the content for a given id.
function show_entry(id)
{
	try{previous=entries_data[id].previous_id} catch(err){}
	try{next=entries_data[id].next_id}catch(err){}
	list_row=document.getElementById('RROW-'+id);
	current_entry_id=id;
	var url='op=view_entry&id='+id;
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

function mark_list_read()
{
	data=$("#entries_form").serialize();
	ndata=data.replace(/id%5B%5D=/g,',')
	data=ndata.replace(/&/g,'')
	alert(data)
	$.ajax({type: "GET",url: '/entry/mark/'+data+'/read', success:function(html){$('#entries_list_div').html(html);$('#menu_status').html('');empty_count();}});
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
// Populates list_div with all entries marked as marked
function marked_entries()
{
	document.getElementById('menu_status').innerHTML='Loading...';
	var url='op=marked_entries';
	var myval='marked_entries';
	$.ajax({
		type: "GET",
		url: "/entries/marked/0/0",
		success: function(html){
			$('#entries_list_div').html(html);
			document.getElementById('menu_status').innerHTML='';
		}
	});
//	var myAjax=new Ajax.Updater('entries_list_div',backend,{method:'get',parameters:url});
}
function showExtendedContentPane(feed_id)
{
	document.getElementById('extendedContentPane').innerHTML='Loading...';
	var url='op=extended_content&id='+feed_id;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#status_div').html(html);}})

	var myAjax=new Ajax.Updater('extendedContentPane',backend,{method:'get',parameters:url});
}
function customize(form)
{
	document.getElementById('menu_status').innerHTML='Loading...';
	var index=form.select.selectedIndex;
	var selvalue=form.select.options[index].value;
	if(selvalue == ''){return;}
	var url='';
	if(selvalue == 'delete'){url='op=delete_feed&id='+current_view_id;}
	else if(selvalue=='default' || selvalue=='link' || selvalue=='extended'||selvalue=='proxy')
		{url='op=update_viewmode&view_mode='+selvalue+'&id='+current_view_id+'&skip_div=1';}
	else
		{url='op=update_category&id='+current_view_id+'&category='+selvalue+'&skip_div=1';}
	$.ajax({type: "GET",url: backend, data:url,success:function(html)
	{
		$('#menu_status').html(html);
		$.ajax({type: "GET",url: backend, data:'op=view_customize_dropdown&id='+current_view_id,success:function(html){$('#customize_dropdown').html(html);}})
	}})
}
function add_feed(form)
{
	$('menu_status').innerHTML='Adding...';
	var newfeed	=form.add_feed_text.value;
	newfeed		=encodeURIComponent(newfeed);
	var url		='op=add_feed&url='+newfeed;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
}
function update_expirey(form)
{
	document.getElementById('menu_status').innerHTML='Submitting...';
	var newexpirey=form.update_feed_expirey.value;
	var url='op=update_expirey&id='+current_view_id+'&expirey='+newexpirey;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
}
function update_autoscroll(form)
{
	document.getElementById('menu_status').innerHTML='Submitting...';
	var newautoscroll=form.update_feed_autoscroll.value;
	var url='op=update_feed_autoscroll&id='+current_view_id+'&autoscroll='+newautoscroll;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
}
function rename_feed(form)
{
	document.getElementById('menu_status').innerHTML='Submitting...';
	var newname=form.rename_feed_text.value;
	var url='op=rename_feed&id='+current_view_id +'&name='+newname;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
}
function rename_category(form)
{
	document.getElementById('menu_status').innerHTML='Submitting...';
	var newname=form.rename_category_text.value;
	var url='op=rename_category&id='+current_view_id +'&name='+newname;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
}
function describe_category(form)
{
	document.getElementById('menu_status').innerHTML='Submitting...';
        var newdesc=form.describe_category_text.value;
	var url='op=describe_category&id='+current_view_id +'&description='+newdesc;
	$.ajax({type: "GET",url: backend, data:url,success:function(html){$('#menu_status').html(html);}})
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
