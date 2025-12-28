import Header from "../../components/Header/Header";

const PrivacyPolicy = () => {
  return (
    <>
      <Header />
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-blue-100 py-12 px-4">
        <div className="max-w-4xl mx-auto bg-white rounded-2xl shadow-xl p-8 border border-blue-100">
          <h1 className="text-3xl font-bold text-blue-900 mb-6">Privacy Policy</h1>
          <p className="text-sm text-zinc-500 mb-8">
            Last updated: {new Date().toLocaleDateString()}
          </p>

          <div className="space-y-6 text-zinc-700">
            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">
                1. Information We Collect
              </h2>
              <p className="mb-2">
                We collect information that you provide directly to us, including:
              </p>
              <ul className="list-disc ml-6 space-y-1">
                <li>Account information (name, email address, organization name)</li>
                <li>Profile information and preferences</li>
                <li>Communications you send to us</li>
                <li>Information you provide when using our services</li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">
                2. How We Use Your Information
              </h2>
              <p className="mb-2">We use the information we collect to:</p>
              <ul className="list-disc ml-6 space-y-1">
                <li>Provide, maintain, and improve our services</li>
                <li>Process transactions and send related information</li>
                <li>Send technical notices, updates, and support messages</li>
                <li>Respond to your comments and questions</li>
                <li>Monitor and analyze trends, usage, and activities</li>
                <li>Detect, prevent, and address technical issues and fraud</li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">3. Information Sharing</h2>
              <p>
                We do not sell, trade, or otherwise transfer your personally identifiable
                information to third parties without your consent, except as described in this
                Privacy Policy or as required by law. We may share information with:
              </p>
              <ul className="list-disc ml-6 space-y-1 mt-2">
                <li>Service providers who assist in our operations</li>
                <li>Professional advisors such as lawyers and accountants</li>
                <li>Law enforcement or regulatory agencies when required by law</li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">4. Data Security</h2>
              <p>
                We implement appropriate technical and organizational measures to protect the
                security of your personal information. However, please note that no method of
                transmission over the Internet or electronic storage is 100% secure, and we cannot
                guarantee absolute security.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">
                5. Authentication Services
              </h2>
              <p>
                We use Auth0 for authentication services. When you create an account or log in,
                Auth0 processes your authentication information in accordance with their privacy
                policy. We recommend reviewing Auth0's privacy policy to understand how they handle
                your data.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">6. Cookies and Tracking</h2>
              <p>
                We use cookies and similar tracking technologies to track activity on our Service
                and hold certain information. Cookies are files with small amounts of data that are
                stored on your device. You can instruct your browser to refuse all cookies or to
                indicate when a cookie is being sent.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">7. Data Retention</h2>
              <p>
                We retain your personal information for as long as necessary to fulfill the purposes
                outlined in this Privacy Policy, unless a longer retention period is required or
                permitted by law.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">8. Your Rights</h2>
              <p className="mb-2">You have the right to:</p>
              <ul className="list-disc ml-6 space-y-1">
                <li>Access and receive a copy of your personal information</li>
                <li>Rectify inaccurate or incomplete personal information</li>
                <li>Request deletion of your personal information</li>
                <li>Object to or restrict processing of your personal information</li>
                <li>Withdraw consent where processing is based on consent</li>
              </ul>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">9. Children's Privacy</h2>
              <p>
                Our Service is not intended for individuals under the age of 13. We do not knowingly
                collect personal information from children under 13. If you become aware that a
                child has provided us with personal information, please contact us.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">
                10. Changes to This Privacy Policy
              </h2>
              <p>
                We may update our Privacy Policy from time to time. We will notify you of any
                changes by posting the new Privacy Policy on this page and updating the "Last
                updated" date.
              </p>
            </section>

            <section>
              <h2 className="text-xl font-semibold text-blue-800 mb-3">11. Contact Us</h2>
              <p>
                If you have any questions about this Privacy Policy, please contact us at{" "}
                <a
                  href="mailto:whubenschmidt@gmail.com"
                  className="text-blue-600 hover:text-blue-800 underline">
                  whubenschmidt@gmail.com
                </a>
              </p>
            </section>
          </div>
        </div>
      </div>
    </>
  );
};

export default PrivacyPolicy;
