document.getElementById('openModal').onclick = function() {
    document.getElementById('feedbackModal').style.display = 'flex';
    history.pushState(null, null, '#feedback-form');
};

window.onpopstate = function() {
    document.getElementById('feedbackModal').style.display = 'none';
};