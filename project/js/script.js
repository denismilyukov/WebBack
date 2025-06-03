const quantityInput = document.getElementById('quantity');
        const serviceTypeRadios = document.getElementsByName('serviceType');
        const optionSelect = document.getElementById('product');
        const propertyCheckbox = document.getElementById('property');
        const totalPriceDisplay = document.getElementById('totalPrice');
        const optionSelectDiv = document.getElementById('optionSelect');
        const propertyCheckboxDiv = document.getElementById('propertyCheckbox');

        const basePrices = {
            1: 175, // цена для типа 1
            2: 100, // цена для типа 2
            3: 50  // цена для типа 3
        };

        function updateOptions() {
            const selectedType = Array.from(serviceTypeRadios).find(r => r.checked).value;

            optionSelectDiv.classList.add('hidden');
            propertyCheckboxDiv.classList.add('hidden');

            if (selectedType == 1) {
                // Тип 1: ничего не показываем
            } else if (selectedType == 2) {
                // Тип 2: показываем только опции
                optionSelectDiv.classList.remove('hidden');
            } else if (selectedType == 3) {
                // Тип 3: показываем только свойство
                propertyCheckboxDiv.classList.remove('hidden');
            }

            calculateTotal();
        }

        function calculateTotal() {
            const quantity = parseInt(quantityInput.value) || 1;
            if (quantity < 0)
                {return;}
            let total = basePrices[Array.from(serviceTypeRadios).find(r => r.checked).value];

            if (optionSelectDiv.classList.contains('hidden') === false) {
                total += parseInt(optionSelect.value) || 0;
            }

            if (propertyCheckbox.checked) {
                total += 100; // стоимость свойства
            }

            totalPriceDisplay.textContent = total*quantity;
        }

        quantityInput.addEventListener('input', calculateTotal);
        optionSelect.addEventListener('change', calculateTotal);
        propertyCheckbox.addEventListener('change', calculateTotal);
        serviceTypeRadios.forEach(radio => {
            radio.addEventListener('change', () => {
                updateOptions();
                calculateTotal();
            });
        });

        // Инициализация
        updateOptions();

const totalSlides = 8;
let curSlide = 0;
let curPage = 1;
let slidesPerPage = 1;
let totalPages = totalSlides - slidesPerPage + 1;

function updateGallery(){
    if (window.innerWidth >= 800){
        slidesPerPage = 3;
    }
    else slidesPerPage = 1;
    totalPages = totalSlides - slidesPerPage + 1;
}


const gallery = document.getElementById('gallery');
const pager = document.getElementById('pager');

function updatePager(){
    pager.textContent = `Фото ${curPage} из ${totalPages}`;
}

function moveSlide(dir){
    curSlide += dir;
    curPage += dir;
    if (curSlide < 0){
        curSlide = 0;
        curPage = 1;
    }
    if (curPage > totalPages){
        curPage = 1;
        curSlide = 0;
    }
    const offset = curSlide * 100 / slidesPerPage;
    gallery.style.transform = `translateX(-${offset}%)`;
    updatePager();
}

updateGallery();
updatePager();

 window.addEventListener('resize', function e(){
    updateGallery();
    updatePager();
});
